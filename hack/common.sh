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

NATIVELB_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")/../"
    pwd
)"

VENDOR_DIR=$NATIVELB_DIR/vendor


NATIVELB_PROVIDER=${NATIVELB_PROVIDER:-k8s-multus-1.11.1}
NATIVELB_NUM_NODES=${NATIVELB_NUM_NODES:-1}

# Use this environment variable to set a custom pkgdir path
# Useful for cross-compilation where the default -pkdir for cross-builds may not be writable
#NATIVELB_GO_BASE_PKGDIR="${GOPATH}/crossbuild-cache-root/"

# If on a developer setup, expose ocp on 8443, so that the openshift web console can be used (the port is important because of auth redirects)
if [ -z "${JOB_NAME}" ]; then
    NATIVELB_PROVIDER_EXTRA_ARGS="${NATIVELB_PROVIDER_EXTRA_ARGS} --ocp-port 8443"
fi

#If run on jenkins, let us create isolated environments based on the job and
# the executor number
provider_prefix=${JOB_NAME:-${NATIVELB_PROVIDER}}${EXECUTOR_NUMBER}
job_prefix=${JOB_NAME:-nativelb}${EXECUTOR_NUMBER}

# Populate an environment variable with the version info needed.
# It should be used for everything which needs a version when building (not generating)
# IMPORTANT:
# RIGHT NOW ONLY RELEVANT FOR BUILDING, GENERATING CODE OUTSIDE OF GIT
# IS NOT NEEDED NOR RECOMMENDED AT THIS STAGE.

function nativelb_version() {
    if [ -n "${NATIVELB_VERSION}" ]; then
        echo ${NATIVELB_VERSION}
    elif [ -d ${NATIVELB_DIR}/.git ]; then
        echo "$(git describe --always --tags)"
    else
        echo "undefined"
    fi
}
NATIVELB_VERSION="$(nativelb_version)"
