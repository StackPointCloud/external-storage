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
	REGISTRY = quay.io/stackpoint/digitalocean-provisioner
endif

ifeq ($(VERSION),)
	VERSION = latest
endif

all build:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o digitalocean-provisioner ./cmd/digitalocean-provisioner
.PHONY: all build

container: build quick-container
.PHONY: container

quick-container:
	cp digitalocean-provisioner deploy/docker/digitalocean-provisioner
	docker build -t $(REGISTRY):$(VERSION) deploy/docker
.PHONY: quick-container

push: container
	docker push $(REGISTRY):$(VERSION)
.PHONY: push

clean:
	rm -f digitalocean-provisioner
	rm -f deploy/docker/digitalocean-provisioner
.PHONY: clean
