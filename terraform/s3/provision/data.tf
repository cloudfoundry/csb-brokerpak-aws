locals {
  ol_configuration_has_retention_mode  = local.ol_configuration_default_retention_mode != null
  ol_configuration_has_retention_days  = local.ol_configuration_default_retention_days != null
  ol_configuration_has_retention_years = local.ol_configuration_default_retention_years != null
  ol_configuration_has_retention       = (local.ol_configuration_has_retention_mode || local.ol_configuration_has_retention_days || local.ol_configuration_has_retention_years)
  ol_configuration_is_enabled          = local.ol_configuration_default_retention_enabled != null ? local.ol_configuration_default_retention_enabled : local.ol_configuration_has_retention
  # When creating a bucket with Object Lock enabled, Amazon S3 automatically enables versioning for the bucket.
  # To avoid differences between the local state and the AWS state, we will enable versioning when enabling Object Lock.
  is_versioning_enabled = local.enable_versioning ? true : local.ol_enabled

  default_kms_key_as_list = try([coalesce(local.sse_default_kms_key_id)], [])
  extra_kms_keys_as_list  = try(split(",", local.sse_extra_kms_key_ids), [])
  sse_all_kms_key_ids     = join(",", compact(distinct(concat(local.default_kms_key_as_list, local.extra_kms_keys_as_list))))
}
