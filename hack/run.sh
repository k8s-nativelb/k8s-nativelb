#!/bin/bash
NATIVELB_PATH="${NATIVELB_PATH:=`pwd`}"
docker run --rm -t --user $(id -u):$(id -g) \
           --network host \
           --volume `pwd`:/go/src/github.com/k8s-nativelb \
           --env KUBECONFIG=/go/src/github.com/k8s-nativelb/cluster/dind-cluster/config \
           --env COVERALLS_TOKEN="${COVERALLS_TOKEN}" \
           --workdir /go/src/github.com/k8s-nativelb/ \
           quay.io/k8s-nativelb/base-image:latest make $@
