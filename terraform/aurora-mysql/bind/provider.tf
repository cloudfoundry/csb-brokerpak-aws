provider "mysql" {
  endpoint = format("%s:%d", var.hostname, var.port)
  username = var.admin_username
  password = var.admin_password
  tls      = true
}

provider "csbmysql" {
  database = var.name
  password = var.admin_password
  username = var.admin_username
  port     = var.port
  host     = var.hostname
}