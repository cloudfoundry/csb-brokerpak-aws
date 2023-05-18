locals {
  port = 1433
  vpc_id = length(var.aws_vpc_id) > 0 ? data.aws_vpc.provided[0].id : data.aws_vpc.default.id
  subnet_ids = length(var.rds_subnet_group) > 0 ? data.aws_subnets.in_subnet_group.ids : data.aws_subnets.in_vpc.ids
  security_group_ids = length(var.rds_vpc_security_group_ids) > 0 ? data.aws_security_groups.provided.ids : [aws_security_group.rds-sg[0].id]
  subnet_group_name = length(var.rds_subnet_group) > 0 ? data.aws_db_subnet_group.provided[0].name : aws_db_subnet_group.rds-private-subnet[0].name
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "in_vpc" {
  filter {
    name   = "vpc-id"
    values = [local.vpc_id]
  }
}

data "aws_vpc" "provided" {
  count = length(var.aws_vpc_id) > 0 ? 1 : 0
  id = var.aws_vpc_id

  lifecycle {
    postcondition {
      condition     = length(var.aws_vpc_id) == 0 || var.aws_vpc_id == self.id
      error_message = "the specified vpc doesn't exist"
    }
  }
}

data "aws_db_subnet_group" "provided" {
  count = length(var.rds_subnet_group) > 0 ? 1 : 0
  name = var.rds_subnet_group

  lifecycle {
    postcondition {
      condition     = length(var.aws_vpc_id) == 0 || var.aws_vpc_id == self.vpc_id
      error_message = "the specified subnet group doesn't exist or doesn't correspond to the specified vpc"
    }
  }
}

data "aws_security_groups" "provided" {
  filter {
    name   = "vpc-id"
    values = [local.vpc_id]
  }
  filter {
    name   = "group-id"
    values = split(",", var.rds_vpc_security_group_ids)
  }

  lifecycle {
    postcondition {
      condition     = length(var.aws_vpc_id) == 0 || (length(self.vpc_ids) == 1 && contains(self.vpc_ids, var.aws_vpc_id))
      error_message = "the specified security groups don't exist or don't correspond to the specified vpc (1)"
    }
    postcondition {
      condition     = length(var.rds_vpc_security_group_ids) == 0 || split(",", var.rds_vpc_security_group_ids) == self.ids
      error_message = "the specified security groups don't exist or don't correspond to the specified vpc (2)"
    }
  }
}

data "aws_subnets" "in_subnet_group" {
  filter {
    name   = "vpc-id"
    values = [local.vpc_id]
  }
  filter {
    name = "subnet-id"
    values = length(var.rds_subnet_group) > 0 ? data.aws_db_subnet_group.provided[0].subnet_ids : []
  }

  lifecycle {
    postcondition {
      condition     = length(var.rds_subnet_group) == 0 || length(self.ids) > 0
      error_message = "the specified subnet group doesn't exists or doesn't contain any valid subnets"
    }
  }
}
