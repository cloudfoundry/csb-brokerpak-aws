output "username" { value = csbpg_binding_user.new_user.username }
output "password" {
  value     = csbpg_binding_user.new_user.password
  sensitive = true
}
output "name" { value = var.name }
output "uri" {
  value = format(
    "postgresql://%s:%s@%s:%d/%s",
    csbpg_binding_user.new_user.username,
    csbpg_binding_user.new_user.password,
    var.reader_endpoint ? var.reader_hostname : var.hostname,
    var.port,
    var.name,
  )
  sensitive = true
}
output "port" { value = var.port }
output "jdbcUrl" {
  value = format(
    "jdbc:postgresql://%s:%d/%s?user=%s\u0026password=%s\u0026ssl=true",
    var.reader_endpoint ? var.reader_hostname : var.hostname,
    var.port,
    var.name,
    csbpg_binding_user.new_user.username,
    csbpg_binding_user.new_user.password,
  )
  sensitive = true
}
