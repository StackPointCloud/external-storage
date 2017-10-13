# Copyright 2017 The Kubernetes Authors.
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

ifeq ($(REGISTRY),)
	REGISTRY = quay.io/stackpoint/profitbricks-provisioner
endif

ifeq ($(VERSION),)
	VERSION = latest
endif

all: build

build: ci
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o profitbricks-provisioner ./cmd/profitbricks-provisioner
.PHONY: all build

container: build quick-container
.PHONY: container

quick-container:
	cp profitbricks-provisioner deploy/docker/profitbricks-provisioner
	docker build -t $(REGISTRY):$(VERSION) deploy/docker
.PHONY: quick-container

ci: govet gofmt test
.PHONY: ci
gofmt:
	ci/gofmt.sh
.PHONY: gofmt

govet:
	go vet $(shell go list ./... | grep -v vendor)
.PHONY: govet

test:
	go test $(shell go list ./... | grep -v vendor)
.PHONY: test

push: container
	docker push $(REGISTRY):$(VERSION)
.PHONY: push

clean:
	rm -f profitbricks-provisioner
	rm -f deploy/docker/profitbricks-provisioner
.PHONY: clean
