provider "csbsqlserver" {
  server   = var.hostname
  port     = var.port
  username = var.admin_username
  password = var.use_managed_admin_password ? local.managed_admin_password : var.admin_password
  database = var.db_name
  encrypt  = "true"
}

provider "aws" {
  region = var.region
}
