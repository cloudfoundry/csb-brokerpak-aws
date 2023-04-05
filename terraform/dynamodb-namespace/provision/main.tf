resource "aws_iam_user" "housekeeping_user" {
  name = format("%s-housekeeping", var.prefix)
}

resource "aws_iam_access_key" "binding_user_key" {
  user = aws_iam_user.housekeeping_user.name
}

resource "aws_iam_user_policy" "binding_policy" {
  user = aws_iam_user.housekeeping_user.name
  policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Sid"    = "PrefixFullAccess",
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
          var.prefix)
        ]
      },
      {
        "Sid"    = "ListTableAccess",
        "Effect" = "Allow",
        "Action" = [
          "dynamodb:ListTables"
        ],
        "Condition" = {},
        "Resource" = [
          format("arn:%s:dynamodb:%s:%s:table/*",
            data.aws_partition.current.partition,
            var.region,
          data.aws_caller_identity.current.account_id)
        ]
      }
    ]
  })
}

resource "aws_iam_access_key" "housekeeping_user_key" {
  user = aws_iam_user.housekeeping_user.name
}

resource "csbdynamodbns_instance" "housekeeping" {
  access_key_id     = aws_iam_access_key.housekeeping_user_key.id
  secret_access_key = aws_iam_access_key.housekeeping_user_key.secret
}