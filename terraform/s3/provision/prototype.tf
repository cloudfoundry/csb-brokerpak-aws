data "terraform_remote_state" "prev_state" {
  backend = "local"
  config = {
    path = "./terraform.tfstate"
  }
  defaults = {
    inputs : {
      ready : false,
      region : null,
      bucket_name : null,
      acl : null,
      labels : {},
      enable_versioning : false,
      ol_enabled : false,
      boc_object_ownership : "BucketOwnerEnforced",
      pab_block_public_acls : null,
      pab_block_public_policy : null,
      pab_ignore_public_acls : null,
      pab_restrict_public_buckets : null,
      sse_default_kms_key_id : null,
      sse_extra_kms_key_ids : null,
      sse_default_algorithm : null,
      sse_bucket_key_enabled : null,
      aws_s3_bucket_object_lock_configuration : null,
      ol_configuration_default_retention_enabled : null,
      ol_configuration_default_retention_mode : null,
      ol_configuration_default_retention_days : null,
      ol_configuration_default_retention_years : null,
      require_tls : false,
    }
  }
}

locals {
  last_inputs       = data.terraform_remote_state.prev_state.outputs.inputs
  inputs            = merge(local.last_inputs, var.inputs)
  unsupported_props = join(",", setsubtract(keys(var.inputs), keys(var.types)))
}

resource "terraform_data" "strongly_typed_inputs" {
  lifecycle {
    precondition {
      condition     = length(local.unsupported_props) == 0
      error_message = "unsupported properties specified as inputs: ${local.unsupported_props}"
    }
  }
}

resource "terraform_data" "prohibit_update" {
  # Don't run prohibit_update during instance creation
  count = local.last_inputs.ready ? 1 : 0

  lifecycle {
    precondition {
      condition     = local.inputs.bucket_name == local.last_inputs.bucket_name
      error_message = "bucket_name can't be modified after creation"
    }

    precondition {
      condition     = local.inputs.acl == local.last_inputs.acl
      error_message = "acl can't be modified after creation"
    }

    precondition {
      condition     = local.inputs.region == local.last_inputs.region
      error_message = "region can't be modified after creation"
    }

    precondition {
      condition     = local.inputs.boc_object_ownership == local.last_inputs.boc_object_ownership
      error_message = "boc_object_ownership can't be modified after creation"
    }

    precondition {
      condition     = local.inputs.ol_enabled == local.last_inputs.ol_enabled
      error_message = "ol_enabled can't be modified after creation"
    }
  }
}

output "inputs" {
  value = merge(local.inputs, { "ready" : true })
}
