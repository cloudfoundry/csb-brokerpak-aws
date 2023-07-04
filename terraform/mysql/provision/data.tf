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
    // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html
    1  = "db.t2.small"
    2  = "db.t3.medium"
    4  = "db.m5.xlarge"
    8  = "db.m5.2xlarge"
    16 = "db.m5.4xlarge"
    32 = "db.m5.8xlarge"
    64 = "db.m5.16xlarge"
  }

  valid_storage_types_for_iops = ["io1", "gp3"]

  port = 3306

  instance_class = length(var.instance_class) == 0 ? local.instance_types[var.cores] : var.instance_class

  subnet_group = length(var.rds_subnet_group) > 0 ? var.rds_subnet_group : aws_db_subnet_group.rds-private-subnet[0].name

  parameter_group_name = length(var.parameter_group_name) == 0 ? format("default.%s%s", var.engine, var.engine_version) : var.parameter_group_name
  log_groups           = var.enable_audit_logging == true ? { "audit" : true } : {}

  max_allocated_storage = (var.storage_autoscale && var.storage_autoscale_limit_gb > var.storage_gb) ? var.storage_autoscale_limit_gb : null

  rds_vpc_security_group_ids = length(var.rds_vpc_security_group_ids) == 0 ? [
    aws_security_group.rds-sg[0].id
  ] : split(",", var.rds_vpc_security_group_ids)

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
}

data "aws_subnets" "all" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.vpc.id]
  }
}
