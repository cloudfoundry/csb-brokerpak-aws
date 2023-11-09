data "terraform_remote_state" "prev_state" {
  backend = "local"
  config = {
    path = "./terraform.tfstate"
  }
  defaults = {
    inputs : {
      ready : false,
    }
  }
}

locals {
  last_inputs = data.terraform_remote_state.prev_state.outputs.inputs
}

resource "terraform_data" "prohibit_update" {
  # Don't run prohibit_update during instance creation
  count = local.last_inputs.ready ? 1 : 0

  lifecycle {
    precondition {
      condition     = var.bucket_name == local.last_inputs.bucket_name
      error_message = "bucket_name can't be modified after creation"
    }

    precondition {
      condition     = var.acl == local.last_inputs.acl
      error_message = "acl can't be modified after creation"
    }

    precondition {
      condition     = var.region == local.last_inputs.region
      error_message = "region can't be modified after creation"
    }

    precondition {
      condition     = var.boc_object_ownership == local.last_inputs.boc_object_ownership
      error_message = "boc_object_ownership can't be modified after creation"
    }

    precondition {
      condition     = var.ol_enabled == local.last_inputs.ol_enabled
      error_message = "ol_enabled can't be modified after creation"
    }
  }
}

output "inputs" {
  value = {
    "ready" : true
    "region" : var.region
    "bucket_name" : var.bucket_name
    "acl" : var.acl
    "labels" : var.labels
    "enable_versioning" : var.enable_versioning
    "ol_enabled" : var.ol_enabled
    "boc_object_ownership" : var.boc_object_ownership

    "pab_block_public_acls" : var.pab_block_public_acls
    "pab_block_public_policy" : var.pab_block_public_policy
    "pab_ignore_public_acls" : var.pab_ignore_public_acls
    "pab_restrict_public_buckets" : var.pab_restrict_public_buckets

    "sse_default_kms_key_id" : var.sse_default_kms_key_id
    "sse_extra_kms_key_ids" : var.sse_extra_kms_key_ids
    "sse_default_algorithm" : var.sse_default_algorithm
    "sse_bucket_key_enabled" : var.sse_bucket_key_enabled

    "ol_configuration_default_retention_enabled" : var.ol_configuration_default_retention_enabled
    "ol_configuration_default_retention_mode" : var.ol_configuration_default_retention_mode
    "ol_configuration_default_retention_days" : var.ol_configuration_default_retention_days
    "ol_configuration_default_retention_years" : var.ol_configuration_default_retention_years

    "require_tls" : var.require_tls
  }
}
