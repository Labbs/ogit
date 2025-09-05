#!/bin/sh
set -e

# Start the migration
# ./app migration -c /config/config.yaml

# Start the server
exec ./app server -c /config/config.yaml