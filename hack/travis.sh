#!/bin/bash

echo $TRAVIS_BRANCH
docker login -u="$DOCKER_USER" -p="$DOCKER_PASS" quay.io
cd ..

make publish
