#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

docker compose down db cache
docker volume rm muzz_database muzz_cache
docker compose --env-file config/default.env up -d db cache
go run "$SCRIPT_DIR/go/migrate/main.go"
