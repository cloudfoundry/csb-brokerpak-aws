output "access_key_id" {
  value     = aws_iam_access_key.bedrock_key.id
  sensitive = true
}

output "secret_access_key" {
  value     = aws_iam_access_key.bedrock_key.secret
  sensitive = true
}

output "region" {
  value = var.region
}

output "models" {
  value = var.available_models
}

output "bedrock_endpoint" {
  value = var.bedrock_endpoint
}

output "guardrail_id" {
  value = var.guardrail_id
}

output "ttl_expires_at" {
  value = var.ttl_expires_at
}
