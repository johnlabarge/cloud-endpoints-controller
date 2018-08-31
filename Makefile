
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
TAG = dev

all: install

install:
	go install

image:
	docker build -t gcr.io/cloud-solutions-group/cloud-endpoints-controller:$(TAG) .

push: image
	docker push gcr.io/cloud-solutions-group/cloud-endpoints-controller:$(TAG)

install-metacontroller:
	helm install --name metacontroller --namespace metacontroller charts/metacontroller

metalogs:
	kubectl -n metacontroller logs -f metacontroller-metacontroller-0

include test.mk