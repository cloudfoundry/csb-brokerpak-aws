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
  cluster_identifier          = var.instance_name
  engine                      = "aurora-mysql"
  engine_version              = var.engine_version
  database_name               = var.db_name
  master_username             = random_string.username.result
  master_password             = random_password.password.result
  port                        = local.port
  db_subnet_group_name        = local.subnet_group
  vpc_security_group_ids      = local.rds_vpc_security_group_ids
  skip_final_snapshot         = true
  allow_major_version_upgrade = var.allow_major_version_upgrade

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
  count                      = var.cluster_instances
  identifier                 = "${var.instance_name}-${count.index}"
  cluster_identifier         = aws_rds_cluster.cluster.id
  instance_class             = local.serverless ? "db.serverless" : "db.r5.large"
  engine                     = aws_rds_cluster.cluster.engine
  engine_version             = aws_rds_cluster.cluster.engine_version
  db_subnet_group_name       = local.subnet_group
  auto_minor_version_upgrade = var.auto_minor_version_upgrade

  lifecycle {
    prevent_destroy = true
  }
}