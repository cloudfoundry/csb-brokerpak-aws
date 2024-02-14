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
  default = "test-user-several-subscribers-standard-queues"
}

variable "user_name_dlq" {
  type    = string
  default = "test-user-several-subscribers-dlq"
}

resource "aws_sqs_queue" "my_dlq" {
  name = "my-dlq-several-subscribers"
}

resource "aws_sqs_queue" "my_queue" {
  name = "my-queue-several-subscribers"
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.my_dlq.arn
    maxReceiveCount     = 5
  })
}

resource "aws_sqs_queue" "my_queue_two" {
  name = "my-queue-two-several-subscribers"
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.my_dlq.arn
    maxReceiveCount     = 5
  })
}


output "my_queue_url" {
  value = aws_sqs_queue.my_queue.url
}

output "my_queue_two_url" {
  value = aws_sqs_queue.my_queue_two.url
}

output "my_dlq_url" {
  value = aws_sqs_queue.my_dlq.url
}

// binding standard queue
resource "aws_iam_user" "user_standard_queues" {
  name = var.user_name
  path = "/cf/"
}

resource "aws_iam_access_key" "access_key_standard_queues" {
  user = aws_iam_user.user_standard_queues.name
}

resource "aws_iam_user_policy" "user_policy_standard_queues" {
  name = format("%s-p", var.user_name)
  user = aws_iam_user.user_standard_queues.name

  policy = data.aws_iam_policy_document.user_policy_standard_queues.json
}

data "aws_iam_policy_document" "user_policy_standard_queues" {
  statement {
    sid = "sqsAccess"
    actions = [
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:PurgeQueue",
      "sqs:GetQueueAttributes"
    ]
    resources = [
      aws_sqs_queue.my_queue.arn,
      aws_sqs_queue.my_queue_two.arn
    ]
  }
}

output "user_access_key_id_standard_queues" {
  value     = aws_iam_access_key.access_key_standard_queues.id
  sensitive = true
}

output "user_secret_access_key_standard_queues" {
  value     = aws_iam_access_key.access_key_standard_queues.secret
  sensitive = true
}


// binding DLQ
resource "aws_iam_user" "user_dlq" {
  name = var.user_name_dlq
  path = "/cf/"
}

resource "aws_iam_access_key" "access_key_dlq" {
  user = aws_iam_user.user_dlq.name
}

resource "aws_iam_user_policy" "user_policy_dlq" {
  name = format("%s-p", var.user_name_dlq)
  user = aws_iam_user.user_dlq.name

  policy = data.aws_iam_policy_document.user_policy_dlq.json
}

data "aws_iam_policy_document" "user_policy_dlq" {
  statement {
    sid = "sqsDLQAllowUserToReceiveDeletePurgeMessagesAndGetAttr"
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

output "user_access_key_id_dlq" {
  value     = aws_iam_access_key.access_key_dlq.id
  sensitive = true
}

output "user_secret_access_key_dlq" {
  value     = aws_iam_access_key.access_key_dlq.secret
  sensitive = true
}



