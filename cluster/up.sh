#!/bin/bash
#
# Copyright 2018 Red Hat, Inc.
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
#

set -e

./cluster/dind-cluster/dind-cluster-v1.13.sh up
docker run -d -p 5000:5000 --rm --network kubeadm-dind-net --name registry registry:2
kubectl config view --raw > ./cluster/dind-cluster/config

docker build -t localhost:5000/k8s-nativelb/nativelb-nginx ./hack/nginx-docker
docker push localhost:5000/k8s-nativelb/nativelb-nginx

docker build -t localhost:5000/k8s-nativelb/nativelb-client ./hack/client-docker
docker push localhost:5000/k8s-nativelb/nativelb-client