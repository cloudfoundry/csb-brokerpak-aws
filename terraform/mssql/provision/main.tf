resource "aws_security_group" "rds-sg" {
  count  = length(var.rds_vpc_security_group_ids) == 0 ? 1 : 0
  name   = format("%s-sg", var.instance_name)
  vpc_id = local.vpc_id
}

resource "aws_db_subnet_group" "rds-private-subnet" {
  count      = length(var.rds_subnet_group) == 0 ? 1 : 0
  name       = format("%s-p-sn", var.instance_name)
  subnet_ids = local.subnet_ids
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

resource "aws_security_group_rule" "mssql_multiaz_udp_egress" {
  count = length(var.rds_vpc_security_group_ids) == 0 && var.multi_az ? 1 : 0
  # This rule is not strictly mandatory since in AWS all egress traffic is allowed by default.
  # Still, I believe it is valuable:
  # - As a safeguard if such behaviour changes.
  # - As documentation of Multi-AZ requirements.
  # https://aws.amazon.com/rds/sqlserver/faqs/#Multi-AZ_instance_ports_requirement
  security_group_id = aws_security_group.rds-sg[0].id
  type              = "egress"
  protocol          = "udp"
  from_port         = 3343
  to_port           = 3343
  cidr_blocks       = ["0.0.0.0/0"]

}

resource "aws_security_group_rule" "mssql_multiaz_udp_ingress" {
  count             = length(var.rds_vpc_security_group_ids) == 0 && var.multi_az ? 1 : 0
  security_group_id = aws_security_group.rds-sg[0].id
  type              = "ingress"
  protocol          = "udp"
  from_port         = 3343
  to_port           = 3343
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "mssql_multiaz_tcp_egress" {
  count = length(var.rds_vpc_security_group_ids) == 0 && var.multi_az ? 1 : 0
  # This rule is not strictly mandatory since in AWS all egress traffic is allowed by default.
  # Still, I believe it is valuable:
  # - As a safeguard if such behaviour changes.
  # - As documentation of Multi-AZ requirements.
  # https://aws.amazon.com/rds/sqlserver/faqs/#Multi-AZ_instance_ports_requirement
  security_group_id = aws_security_group.rds-sg[0].id
  type              = "egress"
  protocol          = "tcp"
  from_port         = 3343
  to_port           = 3343
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "mssql_multiaz_tcp_ingress" {
  count             = length(var.rds_vpc_security_group_ids) == 0 && var.multi_az ? 1 : 0
  type              = "ingress"
  protocol          = "tcp"
  from_port         = 3343
  to_port           = 3343
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.rds-sg[0].id
}

resource "random_string" "username" {
  length  = 16
  special = false
  numeric = false
}

resource "random_password" "password" {
  length  = 32
  special = false
  // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html#RDS_Limits.Constraints
  override_special = "~_-."
}

resource "aws_db_instance" "db_instance" {
  license_model               = "license-included"
  allocated_storage           = var.storage_gb
  max_allocated_storage       = var.max_allocated_storage
  engine                      = var.engine
  engine_version              = var.mssql_version
  instance_class              = var.instance_class
  identifier                  = var.instance_name
  db_name                     = null # Otherwise: Error: InvalidParameterValue: DBName must be null for engine: sqlserver-xx
  username                    = random_string.username.result
  password                    = var.use_managed_admin_password ? null : random_password.password.result
  manage_master_user_password = var.use_managed_admin_password ? true : null
  tags                        = var.labels
  vpc_security_group_ids      = local.security_group_ids
  db_subnet_group_name        = local.subnet_group_name
  option_group_name           = var.option_group_name
  publicly_accessible         = var.publicly_accessible
  apply_immediately           = true
  storage_encrypted           = var.storage_encrypted
  kms_key_id                  = var.kms_key_id == "" ? null : var.kms_key_id
  skip_final_snapshot         = true
  deletion_protection         = var.deletion_protection
  storage_type                = var.storage_type
  iops                        = contains(local.valid_storage_types_for_iops, var.storage_type) ? var.iops : null
  monitoring_interval         = var.monitoring_interval
  monitoring_role_arn         = var.monitoring_role_arn
  multi_az                    = var.multi_az

  parameter_group_name = length(var.parameter_group_name) == 0 ? aws_db_parameter_group.db_parameter_group[0].name : var.parameter_group_name

  backup_retention_period  = var.backup_retention_period
  backup_window            = var.backup_window
  copy_tags_to_snapshot    = var.copy_tags_to_snapshot
  delete_automated_backups = var.delete_automated_backups
  maintenance_window       = local.maintenance_window
  character_set_name       = var.character_set_name

  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_kms_key_id       = var.performance_insights_kms_key_id == "" ? null : var.performance_insights_kms_key_id
  performance_insights_retention_period = var.performance_insights_enabled ? var.performance_insights_retention_period : null

  allow_major_version_upgrade = var.allow_major_version_upgrade
  auto_minor_version_upgrade  = var.auto_minor_version_upgrade

  enabled_cloudwatch_logs_exports = keys(local.log_groups)

  lifecycle {
    prevent_destroy = true
  }

  timeouts {
    create = "60m"
  }
}

resource "aws_secretsmanager_secret_rotation" "secret_manager" {
  # Note that configuring rotation causes the secret to rotate once as soon as you enable rotation.
  # This happens even if the configured rotation is the same as the AWS default e.g. 7 days.
  count     = var.use_managed_admin_password ? 1 : 0
  secret_id = join("", aws_db_instance.db_instance.master_user_secret.*.secret_arn)
  rotation_rules {
    automatically_after_days = var.rotate_admin_password_after
  }

  depends_on = [aws_db_instance.db_instance]
}

resource "aws_db_parameter_group" "db_parameter_group" {
  count  = length(var.parameter_group_name) == 0 ? 1 : 0
  family = local.family
  # The name cannot be repeated. We need `name_prefix` when upgrading major version.
  # See `DBParameterGroupAlreadyExists` error:
  # https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBParameterGroup.html
  name_prefix = format("rds-mssql-%s", var.instance_name)

  parameter {
    name         = "contained database authentication"
    value        = 1
    apply_method = "immediate" // It is the default value, but it is worth being more explicit.
  }

  parameter {
    name         = "rds.force_ssl"
    value        = var.require_ssl ? 1 : 0
    apply_method = "pending-reboot" // MSSQL engine can't apply this parameter without a reboot. Apply type = Static in AWS
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_cloudwatch_log_group" "this" {
  for_each = local.log_groups
  lifecycle {
    create_before_destroy = true
  }
  name              = "/aws/rds/instance/${var.instance_name}/${each.key}"
  retention_in_days = each.value
  kms_key_id        = var.cloudwatch_log_groups_kms_key_id == "" ? null : var.cloudwatch_log_groups_kms_key_id

  tags = var.labels
}
