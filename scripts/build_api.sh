#!/bin/bash

set -e

docker build -t muzz-api --build-arg="GOLANG_VERSION=1.22.4" --build-arg="ALPINE_VERSION=3.20" .