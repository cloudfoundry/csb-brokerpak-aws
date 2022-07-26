locals {
  ol_configuration_has_retention_mode  = var.ol_configuration_default_retention_mode != null
  ol_configuration_has_retention_days  = var.ol_configuration_default_retention_days != null
  ol_configuration_has_retention_years = var.ol_configuration_default_retention_years != null
  ol_configuration_has_retention       = (local.ol_configuration_has_retention_mode || local.ol_configuration_has_retention_days || local.ol_configuration_has_retention_years)
  ol_configuration_is_enabled          = var.ol_configuration_default_retention_enabled != null ? var.ol_configuration_default_retention_enabled : local.ol_configuration_has_retention
  is_versioning_enabled                = var.enable_versioning ? true : var.ol_enabled
}