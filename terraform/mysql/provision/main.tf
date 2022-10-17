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

resource "aws_security_group" "rds-sg" {
  count  = length(var.rds_vpc_security_group_ids) == 0 ? 1 : 0
  name   = format("%s-sg", var.instance_name)
  vpc_id = data.aws_vpc.vpc.id
}

resource "aws_db_subnet_group" "rds-private-subnet" {
  count      = length(var.rds_subnet_group) == 0 ? 1 : 0
  name       = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnets.all.ids
}

resource "aws_security_group_rule" "rds_inbound_access" {
  count             = length(var.rds_vpc_security_group_ids) == 0 ? 1 : 0
  from_port         = local.port
  protocol          = "tcp"
  security_group_id = aws_security_group.rds-sg[0].id
  to_port           = local.port
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "random_string" "username" {
  length  = 16
  special = false
  numeric = false
}

resource "random_password" "password" {
  length  = 32
  special = false
  // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html#RDS_Limits.Constraints
  override_special = "~_-."
}

resource "aws_db_instance" "db_instance" {
  allocated_storage                     = var.storage_gb
  storage_type                          = var.storage_type
  iops                                  = var.storage_type == "io1" ? var.iops : null
  skip_final_snapshot                   = true
  engine                                = var.engine
  engine_version                        = var.engine_version
  instance_class                        = local.instance_class
  identifier                            = var.instance_name
  db_name                               = var.db_name
  username                              = random_string.username.result
  password                              = random_password.password.result
  parameter_group_name                  = local.parameter_group_name
  tags                                  = var.labels
  vpc_security_group_ids                = local.rds_vpc_security_group_ids
  db_subnet_group_name                  = local.subnet_group
  publicly_accessible                   = var.publicly_accessible
  multi_az                              = var.multi_az
  allow_major_version_upgrade           = var.allow_major_version_upgrade
  auto_minor_version_upgrade            = var.auto_minor_version_upgrade
  maintenance_window                    = local.maintenance_window
  apply_immediately                     = true
  max_allocated_storage                 = local.max_allocated_storage
  storage_encrypted                     = var.storage_encrypted
  kms_key_id                            = var.kms_key_id == "" ? null : var.kms_key_id
  deletion_protection                   = var.deletion_protection
  backup_retention_period               = var.backup_retention_period
  backup_window                         = var.backup_window
  copy_tags_to_snapshot                 = var.copy_tags_to_snapshot
  delete_automated_backups              = var.delete_automated_backups
  option_group_name                     = local.option_group_name
  monitoring_interval                   = var.monitoring_interval
  monitoring_role_arn                   = var.monitoring_role_arn
  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_kms_key_id       = var.performance_insights_kms_key_id == "" ? null : var.performance_insights_kms_key_id
  performance_insights_retention_period = var.performance_insights_enabled ? var.performance_insights_retention_period : null

  # Audit Logging
  enabled_cloudwatch_logs_exports = var.enable_audit_logging == true ? ["audit"] : []

  lifecycle {
    prevent_destroy = true
  }

  depends_on = [aws_cloudwatch_log_group.this]

}
locals {
  log_groups = var.enable_audit_logging == true ? { "audit" : true } : {}
}
resource "aws_cloudwatch_log_group" "this" {
  for_each = local.log_groups
  lifecycle {
    create_before_destroy = true
  }
  name              = "/aws/rds/instance/${var.instance_name}/${each.key}"
  retention_in_days = var.cloudwatch_log_group_retention_in_days
  kms_key_id        = var.cloudwatch_log_group_kms_key_id == "" ? null : var.cloudwatch_log_group_kms_key_id

  tags = var.labels

}
