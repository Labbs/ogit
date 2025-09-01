#!/usr/bin/env bash

curl -X POST http://localhost:8080/repo/create -H "Content-Type: application/json" -d '{
  "name": "new-repo"
}'