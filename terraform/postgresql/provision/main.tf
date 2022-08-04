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
  subnet_ids = data.aws_subnet_ids.all.ids
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
  number  = false
}

resource "random_password" "password" {
  length           = 32
  special          = false
  // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html#RDS_Limits.Constraints
  override_special = "~_-."
}

resource "aws_db_instance" "db_instance" {
  allocated_storage           = var.storage_gb
  storage_type                = "gp2"
  skip_final_snapshot         = true
  engine                      = var.engine
  engine_version              = var.engine_version
  instance_class              = local.instance_class
  identifier                  = var.instance_name
  db_name                     = var.db_name
  username                    = random_string.username.result
  password                    = random_password.password.result
  parameter_group_name        = length(var.parameter_group_name) == 0 ? aws_db_parameter_group.db_parameter_group[0].name : var.parameter_group_name
  tags                        = var.labels
  vpc_security_group_ids      = local.rds_vpc_security_group_ids
  db_subnet_group_name        = local.subnet_group
  publicly_accessible         = var.publicly_accessible
  multi_az                    = var.multi_az
  allow_major_version_upgrade = var.allow_major_version_upgrade
  auto_minor_version_upgrade  = var.auto_minor_version_upgrade
  maintenance_window          = var.maintenance_window == "Sun:00:00-Sun:00:00" ? null : var.maintenance_window
  apply_immediately           = true
  max_allocated_storage       = local.max_allocated_storage
  storage_encrypted           = var.storage_encrypted
  deletion_protection         = var.deletion_protection
  backup_retention_period     = var.backup_retention_period
  backup_window               = var.backup_window == "00:00-00:00" ? null : var.backup_window
  copy_tags_to_snapshot       = var.copy_tags_to_snapshot
  delete_automated_backups    = var.delete_automated_backups
  monitoring_interval         = var.monitoring_interval
  monitoring_role_arn         = var.monitoring_role_arn

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_db_parameter_group" "db_parameter_group" {
  count = length(var.parameter_group_name) == 0 ? 1 : 0

  name   = format("rds-pg-%s", var.instance_name)
  family = format("%s%s", var.engine, var.engine_version)

  parameter {
    name  = "rds.force_ssl"
    value = var.require_ssl ? 1 : 0
  }
}