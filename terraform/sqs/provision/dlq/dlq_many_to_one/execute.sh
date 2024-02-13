#!/usr/bin/env bash
set -e

# Define a function to handle errors
handle_error() {
    echo "An error occurred. Exiting..."
    exit 1
}

# Trap any errors and call the handle_error function
trap 'handle_error' ERR

MY_QUEUE_URL=$(terraform output -json my_queue_url | jq -r)
MY_QUEUE_TWO_URL=$(terraform output -json my_queue_two_url | jq -r)
MY_DLQ_URL=$(terraform output -json my_dlq_url | jq -r)
USER_ACCESS_KEY_ID_STANDARD_QUEUES=$(terraform output -json user_access_key_id_standard_queues | jq -r)
USER_SECRET_ACCESS_KEY_STANDARD_QUEUES=$(terraform output -json user_secret_access_key_standard_queues | jq -r)

USER_ACCESS_KEY_ID_DLQ=$(terraform output -json user_access_key_id_dlq | jq -r)
USER_SECRET_ACCESS_KEY_DLQ=$(terraform output -json user_secret_access_key_dlq | jq -r)

export MY_QUEUE_URL
export MY_QUEUE_TWO_URL
export MY_DLQ_URL

export USER_ACCESS_KEY_ID_STANDARD_QUEUES
export USER_SECRET_ACCESS_KEY_STANDARD_QUEUES

export USER_ACCESS_KEY_ID_DLQ
export USER_SECRET_ACCESS_KEY_DLQ

go run main.go