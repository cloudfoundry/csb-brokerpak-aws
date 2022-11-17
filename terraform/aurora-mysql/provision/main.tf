resource "aws_db_subnet_group" "rds_private_subnet" {
  count      = length(var.rds_subnet_group) == 0 ? 1 : 0
  name       = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnets.all.ids
}

resource "aws_security_group" "rds_sg" {
  count  = length(var.rds_vpc_security_group_ids) == 0 ? 1 : 0
  name   = format("%s-sg", var.instance_name)
  vpc_id = data.aws_vpc.vpc.id
}

resource "aws_security_group_rule" "rds_inbound_access" {
  count             = length(var.rds_vpc_security_group_ids) == 0 ? 1 : 0
  protocol          = "tcp"
  security_group_id = aws_security_group.rds_sg[0].id
  from_port         = local.port
  to_port           = local.port
  type              = "ingress"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "random_string" "username" {
  length  = 16
  special = false
  numeric = false
}

resource "random_password" "password" {
  length  = 41 // This is the limit for Aurora/MySQL, we would prefer longer
  special = false
  // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html#RDS_Limits.Constraints
  override_special = "~_-."
}

resource "aws_rds_cluster" "cluster" {
  cluster_identifier              = var.instance_name
  engine                          = "aurora-mysql"
  engine_version                  = var.engine_version
  database_name                   = var.db_name
  tags                            = var.labels
  master_username                 = random_string.username.result
  master_password                 = random_password.password.result
  port                            = local.port
  db_subnet_group_name            = local.subnet_group
  vpc_security_group_ids          = local.rds_vpc_security_group_ids
  skip_final_snapshot             = true
  allow_major_version_upgrade     = var.allow_major_version_upgrade
  backup_retention_period         = var.backup_retention_period
  preferred_backup_window         = var.preferred_backup_window
  copy_tags_to_snapshot           = var.copy_tags_to_snapshot
  deletion_protection             = var.deletion_protection
  db_cluster_parameter_group_name = var.db_cluster_parameter_group_name
  enabled_cloudwatch_logs_exports = var.enable_audit_logging ? ["audit"] : null
  storage_encrypted               = var.storage_encrypted
  kms_key_id                      = var.kms_key_id
  preferred_maintenance_window    = local.preferred_maintenance_window

  dynamic "serverlessv2_scaling_configuration" {
    for_each = local.serverless ? [null] : []
    content {
      min_capacity = var.serverless_min_capacity
      max_capacity = var.serverless_max_capacity
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  count                                 = var.cluster_instances
  identifier                            = "${var.instance_name}-${count.index}"
  cluster_identifier                    = aws_rds_cluster.cluster.id
  tags                                  = var.labels
  instance_class                        = var.instance_class
  engine                                = aws_rds_cluster.cluster.engine
  engine_version                        = aws_rds_cluster.cluster.engine_version
  db_subnet_group_name                  = local.subnet_group
  auto_minor_version_upgrade            = var.auto_minor_version_upgrade
  monitoring_interval                   = var.monitoring_interval
  monitoring_role_arn                   = var.monitoring_role_arn
  performance_insights_enabled          = var.performance_insights_enabled
  performance_insights_kms_key_id       = var.performance_insights_kms_key_id == "" ? null : var.performance_insights_kms_key_id
  performance_insights_retention_period = var.performance_insights_enabled ? var.performance_insights_retention_period : null
  preferred_maintenance_window          = local.preferred_maintenance_window

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_cloudwatch_log_group" "this" {
  for_each = local.log_groups
  lifecycle {
    create_before_destroy = true
  }
  name              = "/aws/rds/cluster/${var.instance_name}/${each.key}"
  retention_in_days = var.cloudwatch_log_group_retention_in_days
  kms_key_id        = var.cloudwatch_log_group_kms_key_id == "" ? null : var.cloudwatch_log_group_kms_key_id

  tags = var.labels
}
