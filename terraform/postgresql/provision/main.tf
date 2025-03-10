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
  count   = length(var.admin_username) == 0 ? 1 : 0
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
  iops                                  = contains(local.valid_storage_types_for_iops, var.storage_type) ? var.iops : null
  skip_final_snapshot                   = true
  engine                                = local.engine
  engine_version                        = var.postgres_version
  instance_class                        = local.instance_class
  identifier                            = var.instance_name
  db_name                               = var.db_name
  username                              = length(var.admin_username) == 0 ? random_string.username[0].result : var.admin_username
  password                              = var.use_managed_admin_password ? null : random_password.password.result
  manage_master_user_password           = var.use_managed_admin_password ? true : null
  parameter_group_name                  = length(var.parameter_group_name) == 0 ? aws_db_parameter_group.db_parameter_group[0].name : var.parameter_group_name
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
  monitoring_interval                   = var.monitoring_interval
  monitoring_role_arn                   = var.monitoring_role_arn
  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_kms_key_id       = var.performance_insights_kms_key_id == "" ? null : var.performance_insights_kms_key_id
  performance_insights_retention_period = var.performance_insights_enabled ? var.performance_insights_retention_period : null

  enabled_cloudwatch_logs_exports = keys(local.log_groups)

  lifecycle {
    prevent_destroy = true
  }

  # dependencies happen prior to resource expansion,
  # so dependencies are always between resource blocks,
  # not between the individual instances of those resource blocks
  depends_on = [aws_cloudwatch_log_group.this]
}

resource "aws_secretsmanager_secret_rotation" "secret_manager" {
  # Note that configuring rotation causes the secret to rotate once as soon as you enable rotation. 
  # This happens even if the configured rotation is the same as the AWS default e.g. 7 days. 
  count     = var.use_managed_admin_password ? 1 : 0
  secret_id = join("", aws_db_instance.db_instance.master_user_secret.*.secret_arn)

  rotation_rules {
    automatically_after_days = var.rotate_admin_password_after
  }

  depends_on = [aws_db_instance.db_instance]
}

resource "aws_db_parameter_group" "db_parameter_group" {
  count  = length(var.parameter_group_name) == 0 ? 1 : 0
  family = format("%s%s", local.engine, local.major_version)
  # The name cannot be repeated. We need `name_prefix` when upgrading major version.
  # See `DBParameterGroupAlreadyExists` error:
  # https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBParameterGroup.html
  name_prefix = format("rds-pg-%s", var.instance_name)

  parameter {
    name         = "rds.force_ssl"
    value        = var.require_ssl ? 1 : 0
    apply_method = "immediate" // It is the default value, but it is worth being more explicit.
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_cloudwatch_log_group" "this" {
  for_each = local.log_groups
  lifecycle {
    create_before_destroy = true
  }
  name              = "/aws/rds/instance/${var.instance_name}/${each.key}"
  retention_in_days = each.value
  kms_key_id        = var.cloudwatch_log_groups_kms_key_id == "" ? null : var.cloudwatch_log_groups_kms_key_id

  tags = var.labels
}