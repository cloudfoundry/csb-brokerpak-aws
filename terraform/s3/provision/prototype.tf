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

  #############################################################################################
  # The following logic allows leaving any property unspecified (or set as null)              #
  # and it will take the same value that was specified for it (even if such value was null)   #
  # ------------------------------------------------------------------------------------------#
  # Only when you want to modify a property you need to specify its new value. The new value  #
  # will be remembered and you won't have to pass it when running `plan`, `apply`, `destroy`  #
  # Sadly, this naive implementation prevents treating `null` as a real and acceptable value. #
  # A more sophisticated implementation may allow mutating an existing property to `null`     #
  #############################################################################################
  region                                     = var.region != null ? var.region : local.last_inputs.region
  bucket_name                                = var.inputs.bucket_name != null ? var.inputs.bucket_name : local.last_inputs.bucket_name
  acl                                        = var.inputs.acl != null ? var.inputs.acl : local.last_inputs.acl
  labels                                     = var.inputs.labels != null ? var.inputs.labels : local.last_inputs.labels
  enable_versioning                          = var.inputs.enable_versioning != null ? var.inputs.enable_versioning : local.last_inputs.enable_versioning
  ol_enabled                                 = var.inputs.ol_enabled != null ? var.inputs.ol_enabled : local.last_inputs.ol_enabled
  boc_object_ownership                       = var.inputs.boc_object_ownership != null ? var.inputs.boc_object_ownership : local.last_inputs.boc_object_ownership
  pab_block_public_acls                      = var.inputs.pab_block_public_acls != null ? var.inputs.pab_block_public_acls : local.last_inputs.pab_block_public_acls
  pab_block_public_policy                    = var.inputs.pab_block_public_policy != null ? var.inputs.pab_block_public_policy : local.last_inputs.pab_block_public_policy
  pab_ignore_public_acls                     = var.inputs.pab_ignore_public_acls != null ? var.inputs.pab_ignore_public_acls : local.last_inputs.pab_ignore_public_acls
  pab_restrict_public_buckets                = var.inputs.pab_restrict_public_buckets != null ? var.inputs.pab_restrict_public_buckets : local.last_inputs.pab_restrict_public_buckets
  sse_default_kms_key_id                     = var.inputs.sse_default_kms_key_id != null ? var.inputs.sse_default_kms_key_id : local.last_inputs.sse_default_kms_key_id
  sse_extra_kms_key_ids                      = var.inputs.sse_extra_kms_key_ids != null ? var.inputs.sse_extra_kms_key_ids : local.last_inputs.sse_extra_kms_key_ids
  sse_default_algorithm                      = var.inputs.sse_default_algorithm != null ? var.inputs.sse_default_algorithm : local.last_inputs.sse_default_algorithm
  sse_bucket_key_enabled                     = var.inputs.sse_bucket_key_enabled != null ? var.inputs.sse_bucket_key_enabled : local.last_inputs.sse_bucket_key_enabled
  ol_configuration_default_retention_enabled = var.inputs.ol_configuration_default_retention_enabled != null ? var.inputs.ol_configuration_default_retention_enabled : local.last_inputs.ol_configuration_default_retention_enabled
  ol_configuration_default_retention_mode    = var.inputs.ol_configuration_default_retention_mode != null ? var.inputs.ol_configuration_default_retention_mode : local.last_inputs.ol_configuration_default_retention_mode
  ol_configuration_default_retention_days    = var.inputs.ol_configuration_default_retention_days != null ? var.inputs.ol_configuration_default_retention_days : local.last_inputs.ol_configuration_default_retention_days
  ol_configuration_default_retention_years   = var.inputs.ol_configuration_default_retention_years != null ? var.inputs.ol_configuration_default_retention_years : local.last_inputs.ol_configuration_default_retention_years
  require_tls                                = var.inputs.require_tls != null ? var.inputs.require_tls : local.last_inputs.require_tls
}

resource "terraform_data" "prohibit_update" {
  # Don't run prohibit_update during instance creation
  count = local.last_inputs.ready ? 1 : 0

  lifecycle {
    precondition {
      condition     = local.bucket_name == local.last_inputs.bucket_name
      error_message = "bucket_name can't be modified after creation"
    }

    precondition {
      condition     = local.acl == local.last_inputs.acl
      error_message = "acl can't be modified after creation"
    }

    precondition {
      condition     = local.region == local.last_inputs.region
      error_message = "region can't be modified after creation"
    }

    precondition {
      condition     = local.boc_object_ownership == local.last_inputs.boc_object_ownership
      error_message = "boc_object_ownership can't be modified after creation"
    }

    precondition {
      condition     = local.ol_enabled == local.last_inputs.ol_enabled
      error_message = "ol_enabled can't be modified after creation"
    }
  }
}

output "inputs" {
  value = {
    "ready" : true
    "region" : local.region
    "bucket_name" : local.bucket_name
    "acl" : local.acl
    "labels" : local.labels
    "enable_versioning" : local.enable_versioning
    "ol_enabled" : local.ol_enabled
    "boc_object_ownership" : local.boc_object_ownership

    "pab_block_public_acls" : local.pab_block_public_acls
    "pab_block_public_policy" : local.pab_block_public_policy
    "pab_ignore_public_acls" : local.pab_ignore_public_acls
    "pab_restrict_public_buckets" : local.pab_restrict_public_buckets

    "sse_default_kms_key_id" : local.sse_default_kms_key_id
    "sse_extra_kms_key_ids" : local.sse_extra_kms_key_ids
    "sse_default_algorithm" : local.sse_default_algorithm
    "sse_bucket_key_enabled" : local.sse_bucket_key_enabled

    "ol_configuration_default_retention_enabled" : local.ol_configuration_default_retention_enabled
    "ol_configuration_default_retention_mode" : local.ol_configuration_default_retention_mode
    "ol_configuration_default_retention_days" : local.ol_configuration_default_retention_days
    "ol_configuration_default_retention_years" : local.ol_configuration_default_retention_years

    "require_tls" : local.require_tls
  }
}
