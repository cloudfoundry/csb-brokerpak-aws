output "region" {
  value = var.region
}

output "instance_name" {
  value = var.instance_name
}

output "cf_provenance_json" {
  value = jsonencode(local.cf_provenance)
}

output "available_models" {
  value = var.models
}

output "resource_tags_json" {
  value = jsonencode(local.common_tags)
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

output "budget_amount" {
  value = var.budget_amount
}

output "budget_enforcement_mode" {
  value = local.budget_enforcement_mode
}

output "ttl_expires_at" {
  value = local.ttl_expires_at
}
