#!/usr/bin/env bash

curl -X POST http://localhost:8080/api/repo -H "Content-Type: application/json" -d '{
  "name": "new-repo"
}'