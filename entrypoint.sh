#!/usr/bin/env bash

set -e

./app migration -c /config/config.yaml
exec ./app server -c /config/config.yaml