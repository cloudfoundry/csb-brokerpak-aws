output "username" { value = random_string.username.result }

output "password" {
  value     = random_password.password.result
  sensitive = true
}

output "port" {  value = local.port }

output "jdbcUrl" {
  value = format(
    "jdbc:sqlserver://%s:%d;database=%s;user=%s;password=%s;loginTimeout=30;%s",
    var.hostname,
    local.port,
    var.db_name,
    random_string.username.result,
    random_password.password.result,
    local.jdbc_tls_string
  )
  sensitive = true
}

output "uri" {
  value = format(
    "mssql://%s:%d/%s?%s",
    var.hostname,
    local.port,
    var.db_name,
    local.uri_tls_string
  )
  sensitive = true
}
