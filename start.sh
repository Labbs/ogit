#!/usr/bin/env bash

set -e

./tmp/main migration -c config.yaml
exec ./tmp/main server -c config.yaml