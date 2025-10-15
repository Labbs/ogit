#!/bin/bash

curl -X POST http://localhost:8080/api/v1/repository -H "Authorization: Bearer $1" -H "Content-Type: application/json" -d '{"name": "my-repo", "description": "My first repository"}'