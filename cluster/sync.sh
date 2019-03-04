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

./cluster/kubectl.sh get pods -n nativelb -o'custom-columns=status:status.podIP,metadata:metadata.name' --no-headers | grep nativelb-agent-cluster | while read -r line
do
    ipaddr=`echo $line | awk '{print $1}'`
    name=`echo $line | awk '{print $2}'`

    # Create agent object.
    cat << EOF | kubectl create -f - > /dev/null 2>&1
apiVersion: k8s.native-lb/v1
kind: Agent
metadata:
  name: $name
  namespace: nativelb
  labels:
    k8s.nativelb.io/cluster: cluster-sample-cluster
spec:
  hostName: $name
  ipAddress: $ipaddr
  port: 8000
  cluster: cluster-sample-cluster

EOF
done
