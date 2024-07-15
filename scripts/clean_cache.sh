#!/bin/bash

set -e

docker compose down cache
docker volume rm muzz_cache
docker compose --env-file config/default.env up -d cache