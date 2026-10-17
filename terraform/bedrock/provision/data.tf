data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

locals {
  model_ids      = jsondecode(var.models)
  ttl_hours      = coalesce(var.ttl_hours, 8)
  ttl_expires_at = timeadd(timestamp(), "${local.ttl_hours}h")
  budget_enforcement_mode = trimspace(var.budget_alert_email) == "" ? "alert-email-optional-unset" : "alert-email-configured"
  cf_context               = try(jsondecode(var.cf_context_json), {})
  cf_originating_identity = try(jsondecode(var.cf_originating_identity_json), {})
  cf_user_id              = try(local.cf_originating_identity.user_id, local.cf_originating_identity.value.user_id, "")

  cf_provenance = {
    cf_organization_guid = try(local.cf_context.organization_guid, "")
    cf_organization_name = try(local.cf_context.organization_name, "")
    cf_space_guid        = try(local.cf_context.space_guid, "")
    cf_space_name        = try(local.cf_context.space_name, "")
    cf_user_id           = local.cf_user_id
  }

  # Build ARNs for each model ID so the IAM policy can reference them directly.
  model_arns = [
    for m in local.model_ids :
    "arn:${data.aws_partition.current.partition}:bedrock:${var.region}::foundation-model/${m}"
  ]

  common_tags = merge(var.labels, local.cf_provenance, {
    TTLExpiry   = local.ttl_expires_at
    ManagedBy   = "cloud-service-broker"
    Environment = "sandbox"
  })
}
