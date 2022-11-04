output "hostname" { value = var.reader_endpoint ? var.reader_hostname : var.hostname }
output "username" { value = csbmysql_binding_user.new_user.username }
output "password" {
  value     = csbmysql_binding_user.new_user.password
  sensitive = true
}
output "database" { value = var.name }
output "uri" {
  value = format(
    "mysql://%s:%s@%s:%d/%s",
    csbmysql_binding_user.new_user.username,
    csbmysql_binding_user.new_user.password,
    var.hostname,
    var.port,
    var.name,
  )
  sensitive = true
}
output "port" { value = var.port }
output "jdbcUrl" {
  value = format(
    "jdbc:mysql://%s:%d/%s?user=%s\u0026password=%s\u0026useSSL=true",
    var.reader_endpoint ? var.reader_hostname : var.hostname,
    var.port,
    var.name,
    csbmysql_binding_user.new_user.username,
    csbmysql_binding_user.new_user.password,
  )
  sensitive = true
}