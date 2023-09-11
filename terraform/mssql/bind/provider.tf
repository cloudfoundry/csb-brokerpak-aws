provider "csbsqlserver" {
  server   = var.hostname
  port     = local.port
  username = var.admin_username
  password = var.admin_password
  database = var.db_name
  encrypt  = "true"
}