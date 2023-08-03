output "username" { value = random_string.username.result }

output "password" {
  value     = random_password.password.result
  sensitive = true
}

output "jdbcUrl" {
  value = format(
    "jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;Encrypt=true;TrustServerCertificate=false;HostNameInCertificate=*.database.windows.net;loginTimeout=30",
    var.hostname,
    local.port,
    var.db_name,
    random_string.username.result,
    random_password.password.result,
  )
  sensitive = true
}

output "uri" {
  value = format(
    "mssql://%s:%d/%s?encrypt=true&TrustServerCertificate=false&HostNameInCertificate=*.database.windows.net",
    var.hostname,
    local.port,
    var.db_name,
  )
  sensitive = true
}
