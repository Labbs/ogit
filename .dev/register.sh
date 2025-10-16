#!/bin/bash

curl -X POST http://localhost:8080/api/v1/auth/register -H "Content-Type: application/json" -d '{"email":"test@example.com", "username":"testuser", "password":"password123"}'