provider "csbsqlserver" {
  server   = var.hostname
  port     = local.port
  username = var.admin_username
  password = var.use_managed_admin_password ? local.managed_admin_password : var.admin_password
  database = var.db_name
  encrypt  = "true"
}

provider "aws" {
  region     = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}