#!/bin/bash

set -e
/go/bin/goveralls -coverprofile=cover.out -service=travis-ci -repotoken ${REPO_TOKEN} -ignore=$(find -regextype posix-egrep -regex ".*generated_mock.*\.go|.*swagger_generated\.go|.*openapi_generated\.go" -printf "%P\n" | paste -d, -s)
