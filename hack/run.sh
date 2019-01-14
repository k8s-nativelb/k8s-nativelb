#!/bin/bash
source hack/common.sh
NATIVELB_PATH="${NATIVELB_PATH:=`pwd`}"

if [ -z "$TRAVIS_REPO_SLUG" ]
then
echo "Mount provider"
VOLUME_COMMAND="${NATIVELB_PATH}/cluster/$NATIVELB_PROVIDER/:$HOME/.kube/"
KUBECONFIG_VALUE="$HOME/.kube/.kubeconfig"
else
echo "Mount travis"
VOLUME_COMMAND="/home/travis/.kube/:$HOME/.kube/"
KUBECONFIG_VALUE="/home/travis/.kube/config"
fi

docker run --rm -t --user $(id -u):$(id -g) \
           --network host \
           --volume `pwd`:/go/src/github.com/k8s-nativelb \
           --volume ${VOLUME_COMMAND} \
           --volume /home/travis/.minikube/:/home/travis/.minikube/ \
           --env KUBECONFIG=${KUBECONFIG_VALUE}\
           --env COVERALLS_TOKEN=${COVERALLS_TOKEN} \
           --workdir /go/src/github.com/k8s-nativelb/ \
           quay.io/k8s-nativelb/base-image:latest make $@
