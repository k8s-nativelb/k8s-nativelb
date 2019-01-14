#!/bin/bash
source hack/common.sh
NATIVELB_PATH="${NATIVELB_PATH:=`pwd`}"
docker run --rm -t --user $(id -u):$(id -g) \
           --network host \
           --volume `pwd`:/go/src/github.com/k8s-nativelb \
           --volume ${NATIVELB_PATH}/cluster/$NATIVELB_PROVIDER/:$HOME/.kube/ \
           --volume $HOME/.minikube:/home/travis/.minikube \
           --env KUBECONFIG=$HOME/.kube/.kubeconfig \
           --env COVERALLS_TOKEN=$COVERALLS_TOKEN \
           --workdir /go/src/github.com/k8s-nativelb/ \
           quay.io/k8s-nativelb/base-image:latest make $@
