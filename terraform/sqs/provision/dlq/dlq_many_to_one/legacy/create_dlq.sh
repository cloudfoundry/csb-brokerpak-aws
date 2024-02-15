#!/usr/bin/env bash

cf create-service aws-sqs standard my_dlq

# AWS Console: get ARN

#arn="arn:aws:sqs:us-west-2:649758297924:cf-e8341513-dfb7-4adc-bd83-e9e1c64c28f9"
cf create-service aws-sqs standard my_standarqueue -c '{"CreateQueue": {"Attributes": { "RedrivePolicy": "{\"deadLetterTargetArn\": \"arn:aws:sqs:us-west-2:649758297924:cf-e8341513-dfb7-4adc-bd83-e9e1c64c28f9\", \"maxReceiveCount\": 5}"}}}'
