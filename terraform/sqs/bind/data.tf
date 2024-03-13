data "aws_iam_policy_document" "user_policy" {
  dynamic "statement" {
    for_each = local.final_policy
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
      "sqs:ListQueueTags"
    ],
    resources : [var.arn]
  }
  dlq_access = {
    sid : "DLQAccess",
    actions : [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:PurgeQueue",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ListQueueTags",
      "sqs:StartMessageMoveTask", // Can start a task but it needs SendMessage on the destination queue
      "sqs:ListDeadLetterSourceQueues"
    ],
    resources : [var.arn]
  }
  standard_dql_access = {
    sid : "sqsAccessDLQ",
    actions : [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ListQueueTags",
      "sqs:StartMessageMoveTask",
      "sqs:CancelMessageMoveTask",
      "sqs:ListMessageMoveTasks"
    ],
    resources : [var.dlq_arn]
  }

  queue_policy_statements = var.dlq ? [local.dlq_access] : [local.standard_access]

  add_dql_access = concat([local.standard_access], [local.standard_dql_access])

  final_policy = length(var.dlq_arn) > 0 ? local.add_dql_access : local.queue_policy_statements
}