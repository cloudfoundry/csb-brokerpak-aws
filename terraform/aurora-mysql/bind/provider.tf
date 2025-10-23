provider "csbmysql" {
  database = var.name
  password = var.use_managed_admin_password ? local.managed_admin_password : var.admin_password
  username = var.admin_username
  port     = var.port
  host     = var.hostname
}

provider "aws" {
  region     = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}