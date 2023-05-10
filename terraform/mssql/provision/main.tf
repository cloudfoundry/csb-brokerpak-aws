resource "aws_security_group" "rds-sg" {
  name   = format("%s-sg", var.instance_name)
  vpc_id = data.aws_vpc.vpc.id
}

resource "aws_db_subnet_group" "rds-private-subnet" {
  name       = format("%s-p-sn", var.instance_name)
  subnet_ids = data.aws_subnets.all.ids
}

resource "aws_security_group_rule" "rds_inbound_access" {
  from_port         = local.port
  protocol          = "tcp"
  security_group_id = aws_security_group.rds-sg.id
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
  length  = 32
  special = false
  // https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html#RDS_Limits.Constraints
  override_special = "~_-."
}

resource "aws_db_instance" "db_instance" {
  license_model          = "license-included"
  allocated_storage      = 20
  engine                 = "sqlserver-ee"
  engine_version         = "15.00.4236.7.v1"
  instance_class         = "db.m6i.xlarge"
  identifier             = var.instance_name
  db_name                = null # Otherwise: Error: InvalidParameterValue: DBName must be null for engine: sqlserver-xx
  username               = random_string.username.result
  password               = random_password.password.result
  tags                   = var.labels
  vpc_security_group_ids = [aws_security_group.rds-sg.id]
  db_subnet_group_name   = aws_db_subnet_group.rds-private-subnet.name
  apply_immediately      = true
  skip_final_snapshot    = true

  lifecycle {
    prevent_destroy = true
  }

  timeouts {
    create = "60m"
  }
}