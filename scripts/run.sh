#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

docker compose --env-file config/default.env --profile api up -d --build --force-recreate
go run "$SCRIPT_DIR/go/migrate/main.go"