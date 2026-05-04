# Per-binding IAM user scoped to the instance Bedrock policy.
resource "aws_iam_user" "bedrock_user" {
  #checkov:skip=CKV_AWS_273: CSB brokerpak pattern requires programmatic IAM users for per-binding access keys returned in VCAP_SERVICES. SSO is incompatible with the OSBAPI bind/unbind lifecycle. Matches S3/RDS/Redis brokerpak precedent. Keys are scoped to sandbox TTL.
  name = var.user_name
  path = "/cf/bedrock/"
}

resource "aws_iam_access_key" "bedrock_key" {
  user = aws_iam_user.bedrock_user.name
}

resource "aws_iam_user_policy_attachment" "bedrock_policy" {
  user       = aws_iam_user.bedrock_user.name
  policy_arn = var.policy_arn
}
