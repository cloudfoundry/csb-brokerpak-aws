provider "csbmysql" {
  database = var.name
  password = var.admin_password
  username = var.admin_username
  port     = var.port
  host     = var.hostname
}