#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

docker compose --env-file docker.env --profile api up -d --no-deps --build api
go run "$SCRIPT_DIR/go/migrate/main.go"