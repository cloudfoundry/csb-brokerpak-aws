resource "terraform_data" "kms_key_was_provided" {
  count = length(var.kms_key_id) > 0 ? 1 : 0

  lifecycle {
    precondition {
      condition     = var.storage_encrypted == true
      error_message = "set `storage_encrypted` to `true` or leave `kms_key_id` field blank"
    }
  }
}

resource "terraform_data" "iops-based-storage-type" {
  count = var.iops != null ? 1 : 0

  lifecycle {
    precondition {
      condition     = contains(["io1", "gp3"], var.storage_type)
      error_message = "set `iops` to `null` or pick a valid `storage_type` such as io1, gp3"
    }
  }
}

resource "terraform_data" "non-iops-based-storage-type" {
  count = var.iops == null ? 1 : 0

  lifecycle {
    precondition {
      condition     = !contains(["io1", "gp3"], var.storage_type)
      error_message = "specify an `iops` value or pick a `storage_type` different than io1, gp3"
    }
  }
}

