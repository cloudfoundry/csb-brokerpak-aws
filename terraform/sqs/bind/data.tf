data "aws_iam_policy_document" "user_policy" {
  dynamic "statement" {
    for_each = local.queue_policy
    content {
      sid       = try(statement.value.sid, null)
      actions   = try(statement.value.actions, null)
      resources = try(statement.value.resources, null)
    }
  }
}

locals {
  standard_access = {
    sid : "sqsAccess",
    actions : [
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:PurgeQueue",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ListQueues",
      "sqs:ListQueueTags",
      "sqs:ListDeadLetterSourceQueues",
    ],
    resources : [var.arn]
  }
  dql_redrive_access = {
    sid : "sqsAccessDLQ",
    actions : [
      "sqs:StartMessageMoveTask",
      "sqs:CancelMessageMoveTask",
      "sqs:ListMessageMoveTasks",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes"
    ],
    resources : [var.dlq_arn]
  }

  queue_policy = length(var.dlq_arn) > 0 ? concat([local.standard_access], [local.dql_redrive_access]) : [local.standard_access]
}