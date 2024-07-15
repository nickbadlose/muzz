#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

seed="${1:-false}"

docker compose down db
docker volume rm muzz_database
docker compose --env-file config/default.env up -d db

if $seed; then
  echo "---                                 ---"
  echo "--- running migrations with seeding ---"
  echo "---                                 ---"
else
  echo "---                    ---"
  echo "--- running migrations ---"
  echo "---                    ---"
fi;

go run "$SCRIPT_DIR/go/migrate/main.go" --seed="$seed"
