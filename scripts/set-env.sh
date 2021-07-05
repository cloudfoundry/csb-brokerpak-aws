#!/usr/bin/env bash

set +x # Hide secrets

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "You must source this script\nsource ${0}" && exit 1

export SECURITY_USER_NAME=brokeruser
export SECURITY_USER_PASSWORD=brokeruserpassword
export DB_HOST=localhost
export DB_USERNAME=broker
export DB_PASSWORD=brokerpass
export DB_NAME=brokerdb
export PORT=8080
