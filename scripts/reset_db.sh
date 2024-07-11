#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

seed="${1:-false}"

docker compose down
docker volume rm muzz_database
docker compose --env-file development.env up -d

if $seed; then
  echo "---                                 ---"
  echo "--- running migrations with seeding ---"
  echo "---                                 ---"
else
  echo "---                    ---"
  echo "--- running migrations ---"
  echo "---                    ---"
fi;

go run "$SCRIPT_DIR/go/migrate.go" --seed="$seed"
