data "aws_iam_policy_document" "user_policy" {
  dynamic "statement" {
    for_each = local.policy_statements
    content {
      sid       = try(statement.value.sid, null)
      actions   = try(statement.value.actions, null)
      resources = [var.arn]
    }
  }
}

locals {
  standard_user = {
    sid : "sqsAccess",
    actions : [
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:PurgeQueue",
      "sqs:GetQueueAttributes"
    ],
  }
  dlq_user = {
    sid : "sqsDLQAllowUserToReceiveDeletePurgeMessagesAndGetAttr",
    actions : [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:PurgeQueue",
      "sqs:GetQueueAttributes"
    ]
  }
  policy_statements = var.dlq ? { policy : local.dlq_user } : { policy : local.standard_user }
}