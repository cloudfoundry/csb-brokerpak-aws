resource "aws_iam_user" "binding_user" {
  name = var.user_name
}

resource "aws_iam_access_key" "binding_user_key" {
  user = aws_iam_user.binding_user.name
}

resource "aws_iam_user_policy" "binding_policy" {
  user   = aws_iam_user.binding_user.name
  policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Sid" = "PrefixFullAccess",
        "Effect" = "Allow",
        "Action" = [
          "dynamodb:*"
        ],
        "Condition" = {},
        "Resource" = [
          format("arn:%s:dynamodb:%s:%s:table/%s*",
            data.aws_partition.current.partition,
            var.region,
            data.aws_caller_identity.current.account_id,
            var.prefix
          )
        ]
      }
    ]
  })
}