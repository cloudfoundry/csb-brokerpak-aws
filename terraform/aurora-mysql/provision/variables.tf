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
variable "db_name" { type = string }
variable "labels" { type = map(any) }
variable "cluster_instances" { type = number }
variable "aws_vpc_id" { type = string }
variable "serverless_min_capacity" { type = number }
variable "serverless_max_capacity" { type = number }
variable "engine_version" { type = string }
variable "rds_subnet_group" { type = string }
variable "rds_vpc_security_group_ids" { type = string }
variable "allow_major_version_upgrade" { type = bool }
variable "auto_minor_version_upgrade" { type = bool }
variable "backup_retention_period" { type = number }
variable "preferred_backup_window" { type = string }
variable "copy_tags_to_snapshot" { type = bool }
variable "deletion_protection" { type = bool }
variable "enable_audit_logging" { type = bool }
variable "db_cluster_parameter_group_name" { type = string }
variable "monitoring_interval" { type = number }
variable "monitoring_role_arn" { type = string }
variable "performance_insights_enabled" { type = bool }
variable "performance_insights_kms_key_id" { type = string }
variable "performance_insights_retention_period" { type = number }
variable "cloudwatch_log_group_retention_in_days" { type = number }
variable "cloudwatch_log_group_kms_key_id" { type = string }
variable "instance_class" { type = string }
variable "storage_encrypted" { type = bool }
variable "kms_key_id" { type = string }
variable "preferred_maintenance_day" { type = string }
variable "preferred_maintenance_start_hour" { type = string }
variable "preferred_maintenance_start_min" { type = string }
variable "preferred_maintenance_end_hour" { type = string }
variable "preferred_maintenance_end_min" { type = string }
variable "admin_username" { type = string }
variable "legacy_instance" { type = bool }
variable "delete_automated_backups" { type = bool }