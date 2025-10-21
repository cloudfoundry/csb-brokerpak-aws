provider "csbmysql" {
  database = var.name
  password = var.use_managed_admin_password ? local.managed_admin_password : var.admin_password
  username = var.admin_username
  port     = var.port
  host     = var.hostname
}