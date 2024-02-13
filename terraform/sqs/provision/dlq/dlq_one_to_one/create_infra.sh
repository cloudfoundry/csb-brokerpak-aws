#!/usr/bin/env bash

set -e

# Define a function to handle errors
handle_error() {
    echo "An error occurred. Exiting..."
    exit 1
}

# Trap any errors and call the handle_error function
trap 'handle_error' ERR

terraform apply -var="aws_secret_access_key=$AWS_SECRET_ACCESS_KEY" -var="aws_access_key_id=$AWS_ACCESS_KEY_ID" -auto-approve


