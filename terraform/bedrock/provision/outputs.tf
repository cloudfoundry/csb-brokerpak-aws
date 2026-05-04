output "region" {
  value = var.region
}

output "available_models" {
  value = var.models
}

output "bedrock_endpoint" {
  value = "https://bedrock-runtime.${var.region}.amazonaws.com"
}

output "policy_arn" {
  value = aws_iam_policy.bedrock_access.arn
}

output "guardrail_id" {
  value = aws_bedrock_guardrail.content_filter.guardrail_id
}

output "guardrail_arn" {
  value = aws_bedrock_guardrail.content_filter.guardrail_arn
}

output "ttl_expires_at" {
  value = local.ttl_expires_at
}
