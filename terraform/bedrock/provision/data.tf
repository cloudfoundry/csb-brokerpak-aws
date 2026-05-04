data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

locals {
  model_ids      = jsondecode(var.models)
  ttl_expires_at = timeadd(timestamp(), "${var.ttl_hours}h")
  budget_enforcement_mode = trimspace(var.budget_alert_email) == "" ? "alert-email-optional-unset" : "alert-email-configured"

  # Build ARNs for each model ID so the IAM policy can reference them directly.
  model_arns = [
    for m in local.model_ids :
    "arn:${data.aws_partition.current.partition}:bedrock:${var.region}::foundation-model/${m}"
  ]

  common_tags = merge(var.labels, {
    TTLExpiry   = local.ttl_expires_at
    ManagedBy   = "cloud-service-broker"
    Environment = "sandbox"
  })
}
