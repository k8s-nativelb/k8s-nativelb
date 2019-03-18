#!/bin/bash
#
# Copyright 2018 k8s-nativelb, Inc.
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

set -ex

registry=localhost:5000
REGISTRY=$registry make docker-build
REGISTRY=$registry make docker-push

./cluster/kubectl.sh delete --ignore-not-found -f ./config/test/k8s-nativelb.yaml
docker rm -f agent-1 | true
docker rm -f agent-2 | true

# Wait until all objects are deleted
until [[ `./cluster/kubectl.sh get ns | grep "nativelb " | wc -l` -eq 0 ]]; do
    sleep 5
done

./cluster/kubectl.sh apply -f ./config/test/k8s-nativelb.yaml
./cluster/kubectl.sh apply -f ./config/develop/

# Make sure all containers are ready
while [ -n "$(./cluster/kubectl.sh get pods --all-namespaces -o'custom-columns=status:status.containerStatuses[*].ready,metadata:metadata.name' --no-headers | grep false)" ]; do
    echo "Waiting for all containers to become ready ..."
    ./cluster/kubectl.sh get pods --all-namespaces -o'custom-columns=status:status.containerStatuses[*].ready,metadata:metadata.name' --no-headers
    sleep 10
done

docker run -d --name agent-1 --rm -v /proc/sys/net/ipv4/ip_nonlocal_bind:/var/proc/sys/net/ipv4/ip_nonlocal_bind \
                              --env CONTROL_IP=10.192.0.11 \
                              --env CONTROL_PORT=8000 \
                              --env CLUSTER_NAME=cluster-external \
                              --cap-add=NET_ADMIN \
                              --cap-add=SYS_MODULE \
                              --ip 10.192.0.11 --network kubeadm-dind-net quay.io/k8s-nativelb/nativelb-agent

docker run -d --name agent-2 --rm -v /proc/sys/net/ipv4/ip_nonlocal_bind:/var/proc/sys/net/ipv4/ip_nonlocal_bind \
                              --env CONTROL_IP=10.192.0.12 \
                              --env CONTROL_PORT=8000 \
                              --env CLUSTER_NAME=cluster-external \
                              --cap-add=NET_ADMIN \
                              --cap-add=SYS_MODULE \
                              --ip 10.192.0.12 --network kubeadm-dind-net quay.io/k8s-nativelb/nativelb-agent

cat << EOF | kubectl create -f - > /dev/null 2>&1
apiVersion: k8s.native-lb/v1
kind: Agent
metadata:
  name: agent-1
  namespace: nativelb
  labels:
    k8s.nativelb.io/cluster: cluster-external
spec:
  hostName: agent-1
  ipAddress: 10.192.0.11
  port: 8000
  cluster: cluster-external
  operational: true

EOF

cat << EOF | kubectl create -f - > /dev/null 2>&1
apiVersion: k8s.native-lb/v1
kind: Agent
metadata:
  name: agent-2
  namespace: nativelb
  labels:
    k8s.nativelb.io/cluster: cluster-external
spec:
  hostName: agent-2
  ipAddress: 10.192.0.12
  port: 8000
  cluster: cluster-external
  operational: true

EOF
