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

SHELL := /bin/bash

#######
# Dev #
#######

# Docker images

registry-all:
	@./dev/registry.sh update

registry-dev:
	@./dev/registry.sh update dev

registry-create:
	@./dev/registry.sh create

# Misc

.PHONY: cli
cli:
	@mkdir -p ./bin
	@GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/cortex ./cli

cli-watch:
	@rerun -watch ./pkg ./cli -ignore ./vendor ./bin -run sh -c "go build -installsuffix cgo -o ./bin/cortex ./cli && echo 'built'"

format:
	@./dev/format.sh

tools:
	@GO111MODULE=off go get -u -v golang.org/x/lint/golint
	@go get -u -v github.com/VojtechVitek/rerun/cmd/rerun
	@pip3 install black

#########
# Tests #
#########

test:
	@./build/test.sh

test-go:
	@./build/test.sh go

test-python:
	@./build/test.sh python

lint:
	@./build/lint.sh

find-missing-version:
	@./build/find-missing-version.sh

test-examples:
	$(MAKE) registry-all
	@./build/test-examples.sh

###############
# CI Commands #
###############

ci-build-images:
	@./build/build-image.sh images/serving-tf serving-tf
	@./build/build-image.sh images/serving-tf-gpu serving-tf-gpu

ci-push-images:
	@./build/push-image.sh serving-tf
	@./build/push-image.sh serving-tf-gpu

ci-build-cli:
	@./build/cli.sh

ci-build-and-upload-cli:
	@./build/cli.sh upload
