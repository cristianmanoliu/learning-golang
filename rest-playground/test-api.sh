#!/usr/bin/env bash

curl -v http://localhost:8080/health

curl -v \
  -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Cristi"}'

curl -v http://localhost:8080/users/1
