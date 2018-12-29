#!/bin/bash

set -e

image="k8s-multus-1.12.2@sha256:4974496beb19a30156c125d7721912b54a705926dcbf66c41f570dca286996ba"

source cluster/ephemeral-provider-common.sh

function up() {
    ${_cli} run $(_add_common_params)

    # Copy k8s config and kubectl
    ${_cli} scp --prefix $provider_prefix /usr/bin/kubectl - >${NATIVELB_PATH}cluster/$NATIVELB_PROVIDER/.kubectl
    chmod u+x ${NATIVELB_PATH}cluster/$NATIVELB_PROVIDER/.kubectl
    ${_cli} scp --prefix $provider_prefix /etc/kubernetes/admin.conf - >${NATIVELB_PATH}cluster/$NATIVELB_PROVIDER/.kubeconfig

    # Set server and disable tls check
    export KUBECONFIG=${NATIVELB_PATH}cluster/$NATIVELB_PROVIDER/.kubeconfig
    ${NATIVELB_PATH}cluster/$NATIVELB_PROVIDER/.kubectl config set-cluster kubernetes --server=https://$(_main_ip):$(_port k8s)
    ${NATIVELB_PATH}cluster/$NATIVELB_PROVIDER/.kubectl config set-cluster kubernetes --insecure-skip-tls-verify=true

    # Make sure that local config is correct
    prepare_config
}