# k8s-nativelb

[![Build Status](https://travis-ci.org/k8s-nativelb/k8s-nativelb.svg?branch=master)](https://travis-ci.org/k8s-nativelb/k8s-nativelb)
[![Licensed under Apache License version 2.0](https://img.shields.io/github/license/k8s-nativelb/k8s-nativelb.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Coverage Status](https://coveralls.io/repos/github/k8s-nativelb/k8s-nativelb/badge.svg?branch=master)](https://coveralls.io/github/k8s-nativelb/k8s-nativelb?branch=master)

**K8S-nativelb** is a loadbalancer add-on for Kubernetes. 
The aim is to provide a LoadBalancer type service for baremetal and clusters without cloud provider.

**Note:** K8S-nativelb is a heavy work in progress.

# Introduction

## Loadbalancer provider for Kubernetes

This project provide the ability to to configure a loadbalancer cluster and deploy agents on top of the kubernetes cluster
to allow the server type LoadBalancer for clusters that doesn't have cloud provider like baremetal deployments.

 
### To start using NativeLB
 

* Deploy k8s-nativelb

 ```yaml
# k8s
kubectl apply -f  
```

* Create a cluster configuration
```bash
cat << EOF | kubectl create -f - > /dev/null 2>&1
apiVersion: k8s.native-lb/v1
kind: Cluster
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
    k8s.nativelb.default: "true"
  name: cluster-internal
  namespace: nativelb
spec:
  default: true
  internal: true
  ipRange: "10.0.0.0/24"
EOF
```

**note** change the ip range if needed

* Deploy the in cluster agent
```bash
cat << EOF | kubectl create -f - > /dev/null 2>&1
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: nativelb-agent
  namespace: nativelb
  labels:
    nativelb.io/cluster: cluster-sample-internal
spec:
  selector:
    matchLabels:
      nativelb.io/daemonset: nativelb-agent
  template:
    metadata:
      labels:
        nativelb.io/daemonset: nativelb-agent
        daemonset.nativelb.io/port: "8000"
    spec:
      hostNetwork: true
      containers:
        - name: agent
          image: quay.io/k8s-nativelb/nativelb-agent:latest
          imagePullPolicy: Always
          securityContext:
            capabilities:
              add: ["NET_ADMIN","SYS_MODULE"]
            privileged: true
          env:
            - name: CONTROL_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
            - name: CONTROL_PORT
              value: "8000"
            - name: CLUSTER_NAME
              value: "cluster-internal"
          volumeMounts:
            - name: bind
              mountPath: /var/proc/sys/net/ipv4/ip_nonlocal_bind
      volumes:
        - name: bind
          hostPath:
            path: /proc/sys/net/ipv4/ip_nonlocal_bind
      terminationGracePeriodSeconds: 15
EOF
```


## Architecture

### Custom Resource Definitions
 * Cluster
 * Agent
 * Farm
 * Server
 * Backend

## License

NativeLB is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.txt).

    Copyright 2016

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
    
[//]: # (Reference links)
    [k8s]: https://kubernetes.io
    [crd]: https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/