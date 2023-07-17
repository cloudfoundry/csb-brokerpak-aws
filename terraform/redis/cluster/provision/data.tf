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

data "aws_vpc" "vpc" {
  default = length(var.aws_vpc_id) == 0
  id      = length(var.aws_vpc_id) == 0 ? null : var.aws_vpc_id
}

locals {
  instance_types = {
    // https://aws.amazon.com/elasticache/pricing
    1   = "cache.t2.small"
    2   = "cache.t3.medium"
    4   = "cache.m5.large"
    8   = "cache.m5.xlarge"
    16  = "cache.r4.xlarge"
    32  = "cache.r4.2xlarge"
    64  = "cache.r4.4xlarge"
    128 = "cache.r4.8xlarge"
    256 = "cache.r5.12xlarge"
  }

  node_type = length(var.node_type) == 0 ? try(local.instance_types[var.cache_size], "") : var.node_type
  port      = 6379

  subnet_group = length(var.elasticache_subnet_group) > 0 ? var.elasticache_subnet_group : aws_elasticache_subnet_group.subnet_group[0].name

  elasticache_vpc_security_group_ids = length(var.elasticache_vpc_security_group_ids) == 0 ? [aws_security_group.sg[0].id] : split(",", var.elasticache_vpc_security_group_ids)

  is_maintenance_window_blank = length(compact([
    var.maintenance_day,
    var.maintenance_start_hour,
    var.maintenance_end_hour,
    var.maintenance_start_min,
    var.maintenance_end_min
  ])) == 0

  maintenance_window = local.is_maintenance_window_blank ? null : format("%s:%s:%s-%s:%s:%s",
    var.maintenance_day,
    var.maintenance_start_hour,
    var.maintenance_start_min,
    var.maintenance_day,
    var.maintenance_end_hour,
    var.maintenance_end_min
  )

  is_backup_window_blank = length(compact([
    var.backup_start_hour,
    var.backup_end_hour,
    var.backup_start_min,
    var.backup_end_min
  ])) == 0

  backup_window = local.is_backup_window_blank ? null : format("%s:%s-%s:%s",
    var.backup_start_hour,
    var.backup_start_min,
    var.backup_end_hour,
    var.backup_end_min
  )
}

data "aws_subnets" "all" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.vpc.id]
  }
}

resource "terraform_data" "major_version" {
  count = var.auto_minor_version_upgrade == true ? 1 : 0
  lifecycle {
    precondition {
      condition     = endswith(var.redis_version, ".x")
      error_message = "A version in the form d.x should be specified if auto_minor_version_upgrade is enabled. For example: 6.x"
    }
  }
}