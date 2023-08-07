resource "terraform_data" "kms_key_was_provided" {
  count = length(var.kms_key_id) > 0 ? 1 : 0

  lifecycle {
    precondition {
      condition     = var.storage_encrypted == true
      error_message = "set `storage_encrypted` to `true` or leave `kms_key_id` field blank"
    }
  }
}

resource "terraform_data" "mssql-express-encryption" {
  count = var.engine == "sqlserver-ex" ? 1 : 0

  lifecycle {
    precondition {
      condition     = var.storage_encrypted == false
      error_message = "sqlserver-ex does not support encryption at rest"
    }
  }
}
