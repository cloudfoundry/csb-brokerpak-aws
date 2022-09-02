provider "csbpg" {
  host            = var.hostname
  port            = var.port
  username        = var.admin_username
  password        = var.admin_password
  database        = var.name
  data_owner_role = "binding_user_group"
  sslmode         = "require"
}
