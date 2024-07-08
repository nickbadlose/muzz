#!/bin/bash

set -e

docker compose down && docker volume rm muzz_database && docker compose --env-file development.env up -d
