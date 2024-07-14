#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

docker compose down
docker volume rm muzz_database muzz_cache
docker compose --env-file development.env up -d
go run "$SCRIPT_DIR/go/migrate/main.go"
