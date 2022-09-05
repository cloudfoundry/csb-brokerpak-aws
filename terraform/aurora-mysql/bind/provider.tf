provider "mysql" {
  endpoint = format("%s:%d", var.hostname, var.port)
  username = var.admin_username
  password = var.admin_password
  tls      = false
}