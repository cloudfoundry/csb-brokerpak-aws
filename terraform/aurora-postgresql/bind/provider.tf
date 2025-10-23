provider "csbpg" {
  host            = var.hostname
  port            = var.port
  username        = var.admin_username
  password        = var.use_managed_admin_password ? local.managed_admin_password : var.admin_password
  database        = var.name
  data_owner_role = "binding_user_group"
  sslmode         = "verify-full"
}

provider "aws" {
  region     = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}
