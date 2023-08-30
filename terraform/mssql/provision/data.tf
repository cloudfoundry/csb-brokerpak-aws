locals {
  port = 1433

  vpc_id             = try(data.aws_vpc.provided[0].id, data.aws_vpc.default[0].id)
  subnet_ids         = try(data.aws_db_subnet_group.provided[0].subnet_ids, data.aws_subnets.in_provided_vpc[0].ids, data.aws_subnets.in_default_vpc[0].ids)
  security_group_ids = try(data.aws_security_groups.provided[0].ids, [aws_security_group.rds-sg[0].id])
  subnet_group_name  = try(data.aws_db_subnet_group.provided[0].name, aws_db_subnet_group.rds-private-subnet[0].name)

  major_version = split(".", data.csbmajorengineversion.major_version_retriever.major_version)[0]
  family        = format("%s-%s.0", var.engine, local.major_version)

  valid_storage_types_for_iops = ["io1", "gp3"]
}

data "aws_vpc" "default" {
  count   = length(var.aws_vpc_id) > 0 ? 0 : 1
  default = true
}

data "aws_subnets" "in_default_vpc" {
  count = length(var.aws_vpc_id) > 0 ? 0 : 1
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default[0].id]
  }

  lifecycle {
    postcondition {
      condition     = length(self.ids) > 0
      error_message = "the default vpc doesn't contain any subnets"
    }
    postcondition {
      condition     = length(self.ids) <= 20
      error_message = "the default vpc contains more than 20 subnets. please specify a different aws_vpc_id or a valid rds_subnet_group containing the desired subnets"
    }
  }
}

data "aws_vpc" "provided" {
  count = length(var.aws_vpc_id) > 0 ? 1 : 0
  id    = var.aws_vpc_id
}

data "aws_subnets" "in_provided_vpc" {
  count = length(var.aws_vpc_id) > 0 ? 1 : 0
  filter {
    name   = "vpc-id"
    values = [var.aws_vpc_id]
  }

  lifecycle {
    postcondition {
      condition     = length(self.ids) > 0
      error_message = "the specified aws_vpc_id doesn't contain any subnets"
    }
    postcondition {
      condition     = length(self.ids) <= 20
      error_message = "the specified aws_vpc_id contains more than 20 subnets. please specify a different aws_vpc_id or a valid rds_subnet_group containing the desired subnets"
    }
  }
}

data "aws_db_subnet_group" "provided" {
  count = length(var.rds_subnet_group) > 0 ? 1 : 0
  name  = var.rds_subnet_group

  lifecycle {
    precondition {
      condition     = length(var.aws_vpc_id) > 0
      error_message = "when specifying rds_subnet_group please specify also the corresponding aws_vpc_id"
    }
    postcondition {
      condition     = var.aws_vpc_id == self.vpc_id
      error_message = "the specified rds_subnet_group doesn't correspond to the specified aws_vpc_id"
    }
    postcondition {
      condition     = length(self.subnet_ids) > 0
      error_message = "the specified rds_subnet_group doesn't contain any subnets"
    }
  }
}

data "aws_security_groups" "provided" {
  count = length(var.rds_vpc_security_group_ids) > 0 ? 1 : 0
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.provided[0].id]
  }
  filter {
    name   = "group-id"
    values = split(",", var.rds_vpc_security_group_ids)
  }

  lifecycle {
    precondition {
      condition     = length(var.aws_vpc_id) > 0
      error_message = "when specifying rds_vpc_security_group_ids please specify also the corresponding aws_vpc_id"
    }
    postcondition {
      condition     = length(distinct(self.vpc_ids)) == 1 && contains(self.vpc_ids, var.aws_vpc_id)
      error_message = "the specified security groups don't exist or don't correspond to the specified vpc (1)"
    }
    postcondition {
      condition     = toset(split(",", var.rds_vpc_security_group_ids)) == toset(self.ids)
      error_message = "the specified security groups don't exist or don't correspond to the specified vpc (2)"
    }
  }
}

data "csbmajorengineversion" "major_version_retriever" {
  engine_version = var.mssql_version
}

data "csbmajorengineversion" "major_version_checker" {
  count          = var.auto_minor_version_upgrade ? 1 : 0
  engine_version = var.mssql_version

  lifecycle {
    postcondition {
      condition     = self.major_version == var.mssql_version
      error_message = "A Major engine version should be specified when auto_minor_version_upgrade is enabled. Expected engine version: ${self.major_version} - got: ${var.mssql_version}"
    }
  }
}
