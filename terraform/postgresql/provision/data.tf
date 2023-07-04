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
    # Enhanced Monitoring is available for all DB instance classes except for the db.m1.small instance class.
    # Consider adding a clarifying comment to improve UX if we add this instance class.
    1  = "db.t2.small"
    2  = "db.t3.medium"
    4  = "db.m5.xlarge"
    8  = "db.m5.2xlarge"
    16 = "db.m5.4xlarge"
    32 = "db.m5.8xlarge"
    64 = "db.m5.16xlarge"
  }

  valid_storage_types_for_iops = ["io1", "gp3"]
  engine        = "postgres"
  major_version = split(".", var.postgres_version)[0]
  port          = 5432

  instance_class = length(var.instance_class) == 0 ? local.instance_types[var.cores] : var.instance_class

  subnet_group = length(var.rds_subnet_group) > 0 ? var.rds_subnet_group : aws_db_subnet_group.rds-private-subnet[0].name

  should_limit_autoscale_storage = var.storage_autoscale && var.storage_autoscale_limit_gb > var.storage_gb
  max_allocated_storage          = local.should_limit_autoscale_storage ? var.storage_autoscale_limit_gb : null

  rds_vpc_security_group_ids = length(var.rds_vpc_security_group_ids) == 0 ? [
    aws_security_group.rds-sg[0].id
  ] : split(",", var.rds_vpc_security_group_ids)

  is_maintenance_window_blank = length(compact([
    var.maintenance_day,
    var.maintenance_start_hour, var.maintenance_end_hour,
    var.maintenance_start_min, var.maintenance_end_min
  ])) == 0

  maintenance_window = local.is_maintenance_window_blank ? null : format("%s:%s:%s-%s:%s:%s",
    var.maintenance_day, var.maintenance_start_hour, var.maintenance_start_min,
  var.maintenance_day, var.maintenance_end_hour, var.maintenance_end_min)

  postgresql_log_group = var.enable_export_postgresql_logs == true ? { postgresql : var.cloudwatch_postgresql_log_group_retention_in_days } : {}
  upgrade_log_group    = var.enable_export_upgrade_logs == true ? { upgrade : var.cloudwatch_upgrade_log_group_retention_in_days } : {}
  log_groups           = merge(local.postgresql_log_group, local.upgrade_log_group)
}

data "aws_subnets" "all" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.vpc.id]
  }
}