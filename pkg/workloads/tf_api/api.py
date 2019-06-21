# Copyright 2019 Cortex Labs, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import sys
import os
import json
import argparse
import tensorflow as tf
import traceback
import time
from flask import Flask, request, jsonify
from flask_api import status
from waitress import serve
import grpc
from tensorflow_serving.apis import (
    predict_pb2,
    get_model_metadata_pb2,
    prediction_service_pb2_grpc,
    get_model_status_pb2,
    model_service_pb2_grpc,
)
from tensorflow_serving.apis import prediction_service_pb2_grpc
from google.protobuf import json_format
from lib.storage import LocalStorage

import consts
from lib import util
from lib.log import get_logger
from lib.exceptions import CortexException, UserRuntimeException, UserException

logger = get_logger()
logger.propagate = False  # prevent double logging (flask modifies root logger)

app = Flask(__name__)

local_cache = {"stub": None, "metadata": None}

DTYPE_TO_VALUE_KEY = {
    "DT_INT32": "intVal",
    "DT_INT64": "int64Val",
    "DT_FLOAT": "floatVal",
    "DT_STRING": "stringVal",
    "DT_BOOL": "boolVal",
    "DT_DOUBLE": "doubleVal",
    "DT_HALF": "halfVal",
    "DT_COMPLEX64": "scomplexVal",
    "DT_COMPLEX128": "dcomplexVal",
}

DTYPE_TO_TF_TYPE = {
    "DT_INT32": tf.int32,
    "DT_INT64": tf.int64,
    "DT_FLOAT": tf.float32,
    "DT_STRING": tf.string,
    "DT_BOOL": tf.bool,
}


def create_raw_prediction_request(sample):
    signature_def = local_cache["metadata"]["signatureDef"]
    signature_key = list(signature_def.keys())[0]
    prediction_request = predict_pb2.PredictRequest()
    prediction_request.model_spec.name = "default"
    prediction_request.model_spec.signature_name = signature_key

    for column_name, value in sample.items():
        shape = [1]
        if util.is_list(value):
            shape = [len(value)]
        sig_type = signature_def[signature_key]["inputs"][column_name]["dtype"]
        tensor_proto = tf.make_tensor_proto([value], dtype=DTYPE_TO_TF_TYPE[sig_type], shape=shape)
        prediction_request.inputs[column_name].CopyFrom(tensor_proto)

    return prediction_request


def create_get_model_metadata_request():
    get_model_metadata_request = get_model_metadata_pb2.GetModelMetadataRequest()
    get_model_metadata_request.model_spec.name = "default"
    get_model_metadata_request.metadata_field.append("signature_def")
    return get_model_metadata_request


def run_get_model_metadata():
    request = create_get_model_metadata_request()
    resp = local_cache["stub"].GetModelMetadata(request, timeout=10.0)
    sigAny = resp.metadata["signature_def"]
    signature_def_map = get_model_metadata_pb2.SignatureDefMap()
    sigAny.Unpack(signature_def_map)
    sigmap = json_format.MessageToDict(signature_def_map)
    return sigmap


def parse_response_proto_raw(response_proto):
    results_dict = json_format.MessageToDict(response_proto)
    outputs = results_dict["outputs"]

    outputs_simplified = {}
    for key in outputs.keys():
        value_key = DTYPE_TO_VALUE_KEY[outputs[key]["dtype"]]
        outputs_simplified[key] = outputs[key][value_key]

    return {"response": outputs_simplified}


def run_predict(sample):
    prediction_request = create_raw_prediction_request(sample)
    response_proto = local_cache["stub"].Predict(prediction_request, timeout=10.0)
    result = parse_response_proto_raw(response_proto)
    util.log_indent("Sample:", indent=4)
    util.log_pretty(sample, indent=6)
    util.log_indent("Prediction:", indent=4)
    util.log_pretty(result, indent=6)
    return result


def prediction_failed(sample, reason=None):
    message = "prediction failed for sample: {}".format(json.dumps(sample))
    if reason:
        message += " ({})".format(reason)

    logger.error(message)
    return message, status.HTTP_406_NOT_ACCEPTABLE


@app.route("/healthz", methods=["GET"])
def health():
    request = get_model_status_pb2.GetModelStatusRequest()
    request.model_spec.name = "default"
    channel = grpc.insecure_channel("localhost:9000")
    stub = model_service_pb2_grpc.ModelServiceStub(channel)
    result = stub.GetModelStatus(request, 10.0)
    model_status = result.model_version_status
    if len(model_status) != 0 and model_status[0].status.error_code == 0:
        return jsonify({"ok": True})

    return (
        "non-zero model version status for model version " + model_status[0].status.error_code,
        status.HTTP_500_INTERNAL_SERVER_ERROR,
    )


@app.route("/", methods=["POST"])
def predict():
    try:
        payload = request.get_json()
    except Exception as e:
        return "Malformed JSON", status.HTTP_400_BAD_REQUEST

    response = {}

    if not util.is_dict(payload) or "samples" not in payload:
        util.log_pretty(payload, logging_func=logger.error)
        return prediction_failed(payload, "top level `samples` key not found in request")

    logger.info("Predicting " + util.pluralize(len(payload["samples"]), "sample", "samples"))

    predictions = []
    samples = payload["samples"]
    if not util.is_list(samples):
        util.log_pretty(samples, logging_func=logger.error)
        return prediction_failed(
            payload, "expected the value of key `samples` to be a list of json objects"
        )

    for i, sample in enumerate(payload["samples"]):
        util.log_indent("sample {}".format(i + 1), 2)
        try:
            result = run_predict(sample)
        except CortexException as e:
            e.wrap("error", "sample {}".format(i + 1))
            logger.error(str(e))
            logger.exception("An error occurred, see `cx logs api {}` for more details.".format(1))
            return prediction_failed(sample, str(e))
        except Exception as e:
            logger.exception("An error occurred, see `cx logs api {}` for more details.".format(2))
            return prediction_failed(sample, str(e))

        predictions.append(result)

    response["predictions"] = predictions

    return jsonify(response)


def start(args):
    logger.info(args)
    logger.info(os.environ)

    channel = grpc.insecure_channel("localhost:" + str(args.tf_serve_port))
    local_cache["stub"] = prediction_service_pb2_grpc.PredictionServiceStub(channel)

    # wait a bit for tf serving to start before querying metadata
    limit = 300
    for i in range(limit):
        try:
            local_cache["metadata"] = run_get_model_metadata()
            break
        except Exception as e:
            if i == limit - 1:
                logger.exception(
                    "An error occurred, see `cx logs api {}` for more details.".format()
                )
                sys.exit(1)

        time.sleep(1)

    logger.info("Serving model")
    serve(app, listen="*:{}".format(args.port))


def main():
    parser = argparse.ArgumentParser()
    na = parser.add_argument_group("required named arguments")
    # na.add_argument("--mode", type=str, required=True, help="Port (on localhost) to use")

    na.add_argument("--port", type=int, required=True, help="Port (on localhost) to use")
    na.add_argument(
        "--tf-serve-port", type=int, required=True, help="Port (on localhost) where TF Serving runs"
    )
    na.add_argument("--model-path", required=True, help="Location of model")
    na.add_argument("--model-dir", required=True, help="Directory to download the model to")
    parser.set_defaults(func=start)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
