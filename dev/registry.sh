#!/bin/bash

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


set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. >/dev/null && pwd)"

source $ROOT/dev/config/build.sh
source $ROOT/dev/util.sh

ecr_logged_in=false

function ecr_login() {
  if [ "$ecr_logged_in" = false ]; then
    blue_echo "Logging in to ECR..."
    ecr_login_command=$(aws ecr get-login --no-include-email --region $REGISTRY_REGION)
    eval $ecr_login_command
    ecr_logged_in=true
    green_echo "Success\n"
  fi
}

function create_registry() {
  aws ecr create-repository --repository-name=cortexlabs/serving-base --region=$REGISTRY_REGION || true
  aws ecr create-repository --repository-name=cortexlabs/serving-tf --region=$REGISTRY_REGION || true
  aws ecr create-repository --repository-name=cortexlabs/serving-tf-gpu --region=$REGISTRY_REGION || true
}

### HELPERS ###

function build() {
  dir=$1
  image=$2
  tag=$3

  blue_echo "Building $image:$tag..."
  echo "docker build $ROOT -f $dir/Dockerfile -t cortexlabs/$image:$tag -t $REGISTRY_URL/cortexlabs/$image:$tag"
  green_echo "Built $image:$tag\n"
}

function build_base() {
  dir=$1
  image=$2

  blue_echo "Building $image..."
  docker build $ROOT -f $dir/Dockerfile -t cortexlabs/$image:latest
  green_echo "Built $image\n"
}

function cache_builder() {
  dir=$1
  image=$2

  blue_echo "Building $image-builder..."
  docker build $ROOT -f $dir/Dockerfile -t cortexlabs/$image-builder:latest --target builder
  green_echo "Built $image-builder\n"
}

function push() {
  ecr_login

  image=$1
  tag=$2

  blue_echo "Pushing $image:$tag..."
  docker push $REGISTRY_URL/cortexlabs/$image:$tag
  green_echo "Pushed $image:$tag\n"
}

function build_and_push() {
  dir=$1
  image=$2
  tag=$3

  build $dir $image $tag
  push $image $tag
}

function cleanup() {
  docker container prune -f
  docker image prune -f
}

cmd=${1:-""}
env=${2:-""}

if [ "$cmd" = "create" ]; then
  create_registry

elif [ "$cmd" = "update" ]; then
  # build_and_push $ROOT/images/serving-base serving-base latest
  # build_and_push $ROOT/images/serving-tf serving-tf latest
  # build_and_push $ROOT/images/serving-tf-gpu serving-tf-gpu latest
  docker build /Users/vishalbollu/src/github.com/cortexlabs/cortex -f /Users/vishalbollu/src/github.com/cortexlabs/cortex/images/serving-base/Dockerfile -t cortexlabs/serving-base:latest
  docker build /Users/vishalbollu/src/github.com/cortexlabs/cortex -f /Users/vishalbollu/src/github.com/cortexlabs/cortex/images/serving-tf/Dockerfile -t cortexlabs/serving-tf:latest
  cleanup
fi
