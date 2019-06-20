#!/bin/bash

# Start the first process

/usr/bin/python3 /src/tf_api/api.py --mode=init --tf-serve-port=$TF_SERVING_PORT --port=$PORT --model-path=$EXTERNAL_MODEL_PATH --model-dir=$MODEL_BASE_PATH

/usr/bin/python3 /src/tf_api/api.py --mode=run --tf-serve-port=$TF_SERVING_PORT --port=$PORT --model-path=$EXTERNAL_MODEL_PATH --model-dir=$MODEL_BASE_PATH &

status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start my_second_process: $status"
  exit $status
fi

tensorflow_model_server --port=$TF_SERVING_PORT --model_base_path=$MODEL_BASE_PATH &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start my_first_process: $status"
  exit $status
fi

while sleep 60; do
  ps aux |grep tensorflow_model_server |grep -q -v grep
  PROCESS_1_STATUS=$?
  ps aux |grep tf_api/api.py |grep -q -v grep
  PROCESS_2_STATUS=$?
  # If the greps above find anything, they exit with 0 status
  # If they are not both 0, then something is wrong
  if [ $PROCESS_1_STATUS -ne 0 -o $PROCESS_2_STATUS -ne 0 ]; then
    echo "One of the processes has already exited."
    exit 1
  fi
done
