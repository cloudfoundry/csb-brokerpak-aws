variable "aws_access_key_id" {
  type      = string
  sensitive = true
}
variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
variable "region" { type = string }
variable "instance_name" { type = string }
variable "storage_encrypted" { type = bool }
variable "kms_key_id" { type = string }
variable "db_name" { type = string }
variable "labels" { type = map(any) }
variable "instance_class" { type = string }
variable "aws_vpc_id" { type = string }
variable "rds_subnet_group" { type = string }
variable "rds_vpc_security_group_ids" { type = string }
variable "mssql_version" { type = string }
variable "engine" { type = string }
variable "storage_gb" { type = number }
variable "deletion_protection" { type = bool }
variable "max_allocated_storage" {
  type     = number
  nullable = true
}
variable "storage_type" { type = string }
variable "iops" { type = number }
variable "publicly_accessible" { type = bool }
variable "monitoring_interval" { type = number }
variable "monitoring_role_arn" { type = string }
variable "option_group_name" { type = string }
variable "parameter_group_name" { type = string }
variable "backup_retention_period" { type = string }
variable "backup_window" { type = string }
variable "copy_tags_to_snapshot" { type = bool }
variable "delete_automated_backups" { type = bool }
variable "maintenance_day" { type = string }
variable "maintenance_start_hour" { type = string }
variable "maintenance_start_min" { type = string }
variable "maintenance_end_hour" { type = string }
variable "maintenance_end_min" { type = string }
variable "allow_major_version_upgrade" { type = bool }
variable "auto_minor_version_upgrade" { type = bool }
variable "require_ssl" { type = bool }
variable "character_set_name" { type = string }
variable "performance_insights_enabled" { type = bool }
variable "performance_insights_kms_key_id" { type = string }
variable "performance_insights_retention_period" { type = number }
variable "enable_export_agent_logs" { type = bool }
variable "enable_export_error_logs" { type = bool }
variable "cloudwatch_agent_log_group_retention_in_days" { type = number }
variable "cloudwatch_error_log_group_retention_in_days" { type = number }
variable "cloudwatch_log_groups_kms_key_id" { type = string }
variable "multi_az" { type = bool }
variable "use_managed_admin_password" { type = bool }
variable "rotate_admin_password_after" { type = number }