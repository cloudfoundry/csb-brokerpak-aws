locals {
  // Convert the default KMS master key ID to a list if it's not empty, or an empty list otherwise.
  default_kms_key_id_as_list = var.kms_master_key_id != "" ? [var.kms_master_key_id] : []

  // Split the extra KMS key IDs into a list if not empty, or an empty list otherwise.
  kms_extra_key_ids_as_list = var.kms_extra_key_ids != "" ? split(",", var.kms_extra_key_ids) : []

  // Combine the default and extra KMS key IDs into a single list, ensuring distinct values and removing any empty elements,
  // then join them into a single comma-separated string.
  kms_all_key_ids = join(",", compact(distinct(concat(local.default_kms_key_id_as_list, local.kms_extra_key_ids_as_list))))
}