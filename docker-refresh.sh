#!/usr/bin/env bash

set -e

docker rm -f zk-playground pg-playground kafka-playground 2>/dev/null || true
docker compose down --remove-orphans
docker compose up -d
