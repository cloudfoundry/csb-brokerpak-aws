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

  port = 5432

  instance_class = length(var.instance_class) == 0 ? local.instance_types[var.cores] : var.instance_class

  subnet_group = length(var.rds_subnet_group) > 0 ? var.rds_subnet_group : aws_db_subnet_group.rds-private-subnet[0].name

  max_allocated_storage = (var.storage_autoscale && var.storage_autoscale_limit_gb > var.storage_gb) ? var.storage_autoscale_limit_gb : null

  rds_vpc_security_group_ids = length(var.rds_vpc_security_group_ids) == 0 ? [aws_security_group.rds-sg[0].id] : split(",", var.rds_vpc_security_group_ids)
}

data "aws_subnet_ids" "all" {
  vpc_id = data.aws_vpc.vpc.id
}
