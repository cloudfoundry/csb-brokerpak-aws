data "aws_iam_policy_document" "user_policy" {
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
      var.arn
    ]
  }
}