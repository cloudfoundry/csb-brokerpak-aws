terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

provider "aws" {
  region     = "us-west-2"
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

variable "user_name" {
  type    = string
  default = "test-user-dlq"
}

resource "aws_sqs_queue" "my_dlq" {
  name = "my-dlq"
}

resource "aws_sqs_queue" "my_queue" {
  name           = "my-queue"
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.my_dlq.arn
    maxReceiveCount     = 5
  })
}

output "my_queue_url" {
  value = aws_sqs_queue.my_queue.url
}

output "my_dlq_url" {
  value = aws_sqs_queue.my_dlq.url
}

// binding
resource "aws_iam_user" "user" {
  name = var.user_name
  path = "/cf/"
}

resource "aws_iam_access_key" "access_key" {
  user = aws_iam_user.user.name
}

resource "aws_iam_user_policy" "user_policy" {
  name = format("%s-p", var.user_name)
  user = aws_iam_user.user.name

  policy = data.aws_iam_policy_document.user_policy.json
}

data "aws_iam_policy_document" "user_policy" {
  statement {
    sid     = "sqsAccess"
    actions = [
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:PurgeQueue",
      "sqs:GetQueueAttributes"
    ]
    resources = [
      aws_sqs_queue.my_queue.arn
    ]
  }

  statement {
    sid     = "sqsDLQAllowUserToReceiveDeletePurgeMessagesAndGetAttr"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:PurgeQueue",
      "sqs:GetQueueAttributes"
    ]
    resources = [
      aws_sqs_queue.my_dlq.arn
    ]
  }
}

output "user_access_key_id" {
  value     = aws_iam_access_key.access_key.id
  sensitive = true
}
output "user_secret_access_key" {
  value     = aws_iam_access_key.access_key.secret
  sensitive = true
}

