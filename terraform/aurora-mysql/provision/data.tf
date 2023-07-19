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
  engine     = "aurora-mysql"
  port       = 3306
  serverless = var.serverless_max_capacity != null || var.serverless_min_capacity != null

  rds_vpc_security_group_ids = length(var.rds_vpc_security_group_ids) == 0 ? [aws_security_group.rds_sg[0].id] : split(",", var.rds_vpc_security_group_ids)
  subnet_group               = length(var.rds_subnet_group) > 0 ? var.rds_subnet_group : aws_db_subnet_group.rds_private_subnet[0].name

  log_groups = var.enable_audit_logging == true ? { "audit" : true } : {}

  is_maintenance_window_blank = length(compact([
    var.preferred_maintenance_day,
    var.preferred_maintenance_start_hour,
    var.preferred_maintenance_end_hour,
    var.preferred_maintenance_start_min,
    var.preferred_maintenance_end_min
  ])) == 0

  preferred_maintenance_window = local.is_maintenance_window_blank ? null : format("%s:%s:%s-%s:%s:%s",
    var.preferred_maintenance_day,
    var.preferred_maintenance_start_hour,
    var.preferred_maintenance_start_min,
    var.preferred_maintenance_day,
    var.preferred_maintenance_end_hour,
    var.preferred_maintenance_end_min
  )
}

data "csbmajorengineversion" "major_version_checker" {
  count          = var.auto_minor_version_upgrade ? 1 : 0
  engine_version = var.engine_version

  lifecycle {
    postcondition {
      condition     = self.major_version == var.engine_version
      error_message = "A Major engine version should be specified when auto_minor_version_upgrade is enabled. Expected engine version: ${self.major_version} - got: ${var.engine_version}"
    }
  }
}