# Copyright 2020 Pivotal Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

variable "aws_access_key_id" {
  type      = string
  sensitive = true
}
variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
variable "region" { type = string }
variable "cores" { type = number }
variable "instance_name" { type = string }
variable "db_name" { type = string }
variable "labels" { type = map(any) }
variable "storage_gb" { type = number }
variable "storage_type" { type = string }
variable "iops" { type = number }
variable "publicly_accessible" { type = bool }
variable "multi_az" { type = bool }
variable "instance_class" { type = string }
variable "postgres_version" { type = string }
variable "use_managed_admin_password" { type = bool }
variable "rotate_admin_password_after" { type = number }
variable "aws_vpc_id" { type = string }
variable "storage_autoscale" { type = bool }
variable "storage_autoscale_limit_gb" { type = number }
variable "storage_encrypted" { type = bool }
variable "kms_key_id" { type = string }
variable "parameter_group_name" { type = string }
variable "rds_subnet_group" { type = string }
variable "rds_vpc_security_group_ids" { type = string }
variable "allow_major_version_upgrade" { type = bool }
variable "auto_minor_version_upgrade" { type = bool }
variable "maintenance_day" { type = string }
variable "maintenance_start_hour" { type = string }
variable "maintenance_start_min" { type = string }
variable "maintenance_end_hour" { type = string }
variable "maintenance_end_min" { type = string }
variable "require_ssl" { type = bool }
variable "provider_verify_certificate" { type = bool }
variable "deletion_protection" { type = bool }
variable "backup_retention_period" { type = number }
variable "backup_window" { type = string }
variable "copy_tags_to_snapshot" { type = bool }
variable "delete_automated_backups" { type = bool }
variable "monitoring_interval" { type = number }
variable "monitoring_role_arn" { type = string }
variable "performance_insights_enabled" { type = bool }
variable "performance_insights_kms_key_id" { type = string }
variable "performance_insights_retention_period" { type = number }
variable "enable_export_postgresql_logs" { type = bool }
variable "enable_export_upgrade_logs" { type = bool }
variable "cloudwatch_postgresql_log_group_retention_in_days" { type = number }
variable "cloudwatch_upgrade_log_group_retention_in_days" { type = number }
variable "cloudwatch_log_groups_kms_key_id" { type = string }
variable "admin_username" { type = string }
