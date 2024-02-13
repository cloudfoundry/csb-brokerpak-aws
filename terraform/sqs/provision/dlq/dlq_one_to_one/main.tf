terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}

variable "aws_access_key_id" {
  type      = string
  sensitive = true
}
variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}

resource "aws_sqs_queue" "my_dlq" {
  name = "my-dlq"
}

resource "aws_sqs_queue" "my_queue" {
  name                        = "my-queue"
  redrive_policy              = jsonencode({
    deadLetterTargetArn       = aws_sqs_queue.my_dlq.arn
    maxReceiveCount           = 5
  })
}

output "my_queue_url" {
  value = aws_sqs_queue.my_queue.url
}

output "my_dlq_url" {
  value = aws_sqs_queue.my_dlq.url
}


