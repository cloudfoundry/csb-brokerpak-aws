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

  kms_statement = {
    sid : "kmsAccess",
    actions : [
      "kms:GenerateDataKey",
      "kms:Decrypt"
    ]
    resources = [for key in data.aws_kms_key.customer_provided_keys : key.arn]
  }

  key_ids_list = compact(split(",", var.kms_all_key_ids))
  has_key_ids  = length(local.key_ids_list) != 0

  queue_policy = concat(
    [local.standard_access],
    length(var.dlq_arn) > 0 ? [local.dql_redrive_access] : [],
    local.has_key_ids ? [local.kms_statement] : []
  )
}

data "aws_iam_policy_document" "user_policy" {
  dynamic "statement" {
    for_each = local.queue_policy
    content {
      sid       = statement.value.sid
      actions   = statement.value.actions
      resources = statement.value.resources
    }
  }
}

data "aws_kms_key" "customer_provided_keys" {
  count  = local.has_key_ids ? length(local.key_ids_list) : 0
  key_id = local.key_ids_list[count.index]
}