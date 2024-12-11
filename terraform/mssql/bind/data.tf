data "aws_secretsmanager_secret_version" "secret-version" {
  count     = var.use_managed_admin_password ? 1 : 0
  secret_id = var.managed_admin_credentials_arn
}

locals {
  jdbc_tls_string        = (var.require_ssl ? format("encrypt=true;trustServerCertificate=false;hostNameInCertificate=%s", var.hostname) : "encrypt=false")
  uri_tls_string         = (var.require_ssl ? format("encrypt=true&TrustServerCertificate=false&HostNameInCertificate=%s", var.hostname) : "encrypt=false")
  managed_admin_creds    = var.use_managed_admin_password ? jsondecode(data.aws_secretsmanager_secret_version.secret-version[0].secret_string) : {}
  managed_admin_password = var.use_managed_admin_password ? local.managed_admin_creds.password : ""
}