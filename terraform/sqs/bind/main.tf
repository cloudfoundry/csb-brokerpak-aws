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
