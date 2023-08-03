resource "terraform_data" "kms_key_was_not_provided" {
  count = length(var.kms_key_id) == 0 ? 1 : 0

  lifecycle {
    precondition {
      condition     = var.storage_encrypted == false
      error_message = "set `storage_encrypted` to `false` or provide a valid `kms_key_id`"
    }
  }
}

resource "terraform_data" "kms_key_was_provided" {
  count = length(var.kms_key_id) > 0 ? 1 : 0

  lifecycle {
    precondition {
      condition     = var.storage_encrypted == true
      error_message = "set `storage_encrypted` to `true` or leave `kms_key_id` field blank"
    }
  }
}
