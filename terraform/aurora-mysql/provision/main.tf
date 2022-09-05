resource "aws_security_group" "rds_sg" {
  name   = format("%s-sg", var.instance_name)
  vpc_id = data.aws_vpc.vpc.id
}

resource "aws_db_subnet_group" "rds_private_subnet" {
  name       = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnets.all.ids
}

resource "aws_security_group_rule" "rds_inbound_access" {
  from_port         = local.port
  protocol          = "tcp"
  security_group_id = aws_security_group.rds_sg.id
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
  cluster_identifier     = var.instance_name
  engine                 = "aurora-mysql"
  database_name          = "auroradb"
  master_username        = random_string.username.result
  master_password        = random_password.password.result
  port                   = local.port
  db_subnet_group_name   = aws_db_subnet_group.rds_private_subnet.name
  vpc_security_group_ids = [aws_security_group.rds_sg.id]
  skip_final_snapshot    = true

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_rds_cluster_instance" "cluster_instances" {
  count                = 1
  identifier           = "${var.instance_name}-${count.index}"
  cluster_identifier   = aws_rds_cluster.cluster.id
  instance_class       = "db.r5.large"
  engine               = aws_rds_cluster.cluster.engine
  engine_version       = aws_rds_cluster.cluster.engine_version
  db_subnet_group_name = aws_db_subnet_group.rds_private_subnet.name

  lifecycle {
    prevent_destroy = true
  }
}