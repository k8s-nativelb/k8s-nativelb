#!/bin/bash
source hack/common.sh
NATIVELB_PATH="${NATIVELB_PATH:=`pwd`}"
docker run --rm -t --volume `pwd`:/go/src/github.com/k8s-nativelb  --volume ${NATIVELB_PATH}/cluster/$NATIVELB_PROVIDER/:/root/.kube/ --volume $HOME/.minikube:/home/travis/.minikube --env KUBECONFIG=/root/.kube/.kubeconfig --workdir /go/src/github.com/k8s-nativelb/ quay.io/k8s-nativelb/base-image:latest make $@
