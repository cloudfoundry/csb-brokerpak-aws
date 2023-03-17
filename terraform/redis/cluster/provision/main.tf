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

resource "aws_security_group" "sg" {
  count  = length(var.elasticache_vpc_security_group_ids) == 0 ? 1 : 0
  name   = format("%s-sg", var.instance_name)
  vpc_id = data.aws_vpc.vpc.id
}

resource "aws_elasticache_subnet_group" "subnet_group" {
  count      = length(var.elasticache_subnet_group) == 0 ? 1 : 0
  name       = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnets.all.ids
}

resource "aws_security_group_rule" "inbound_access" {
  count             = length(var.elasticache_vpc_security_group_ids) == 0 ? 1 : 0
  from_port         = local.port
  protocol          = "tcp"
  security_group_id = aws_security_group.sg[0].id
  to_port           = local.port
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "random_password" "auth_token" {
  length = 64
  // https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/auth.html
  override_special = "!&#$^<>-"
  min_upper        = 2
  min_lower        = 2
  min_special      = 2
}

resource "aws_elasticache_replication_group" "redis" {
  automatic_failover_enabled = var.node_count > 1
  multi_az_enabled           = var.multi_az_enabled
  replication_group_id       = var.instance_name
  description                = format("%s redis", var.instance_name)
  node_type                  = local.node_type
  num_cache_clusters         = var.node_count
  engine_version             = var.redis_version
  port                       = local.port
  tags                       = var.labels
  security_group_ids         = local.elasticache_vpc_security_group_ids
  subnet_group_name          = local.subnet_group
  transit_encryption_enabled = true
  auth_token                 = random_password.auth_token.result
  apply_immediately          = true
  at_rest_encryption_enabled = var.at_rest_encryption_enabled
  kms_key_id                 = var.kms_key_id
  maintenance_window         = local.maintenance_window
  data_tiering_enabled       = var.data_tiering_enabled
  snapshot_retention_limit   = var.backup_retention_limit
  final_snapshot_identifier  = var.final_backup_identifier
  snapshot_name              = var.backup_name

  lifecycle {
    prevent_destroy = true
  }
}
