output "access_key_id" {
  value     = aws_iam_access_key.bedrock_key.id
  sensitive = true
}

output "instance_name" {
  value = var.instance_name
}

output "resource_tags_json" {
  value = jsonencode(merge(
    jsondecode(var.resource_tags_json),
    var.cf_app_guid == "" ? {} : { cf_app_guid = var.cf_app_guid },
  ))
}

output "binding_provenance_json" {
  value = jsonencode({
    cf_organization_guid = try(jsondecode(var.cf_context_json).organization_guid, "")
    cf_organization_name = try(jsondecode(var.cf_context_json).organization_name, "")
    cf_space_guid        = try(jsondecode(var.cf_context_json).space_guid, "")
    cf_space_name        = try(jsondecode(var.cf_context_json).space_name, "")
    cf_user_id           = try(jsondecode(var.cf_originating_identity_json).user_id, jsondecode(var.cf_originating_identity_json).value.user_id, "")
    cf_app_guid          = var.cf_app_guid
  })
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

output "normalized_binding_json" {
  value = jsonencode({
    version            = "v1"
    provider           = "aws"
    provisioner_family = "aws_bedrock_identity"
    connection_type    = "runtime"
    endpoint = {
      base_url    = var.bedrock_endpoint
      region      = var.region
      api_version = null
    }
    access = {
      mode       = "aws_sigv4"
      expires_at = null
    }
    grant = {
      kind                 = "scoped_key"
      least_privilege_unit = "model"
      allowed_models       = try(jsondecode(var.available_models), [])
    }
    credential = {
      format = "aws_temp_creds"
      inline = {
        access_key_id     = aws_iam_access_key.bedrock_key.id
        secret_access_key = aws_iam_access_key.bedrock_key.secret
      }
      secret_ref = null
    }
  })
  sensitive = true
}
