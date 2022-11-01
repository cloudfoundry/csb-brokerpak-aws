data "aws_vpc" "vpc" {
  default = length(var.aws_vpc_id) == 0
  id      = length(var.aws_vpc_id) == 0 ? null : var.aws_vpc_id
}

data "aws_subnets" "all" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.vpc.id]
  }
}

locals {
  port       = 3306
  serverless = var.serverless_max_capacity != null || var.serverless_min_capacity != null

  rds_vpc_security_group_ids = length(var.rds_vpc_security_group_ids) == 0 ? [aws_security_group.rds_sg[0].id] : split(",", var.rds_vpc_security_group_ids)
  subnet_group               = length(var.rds_subnet_group) > 0 ? var.rds_subnet_group : aws_db_subnet_group.rds_private_subnet[0].name
}