output "hostname" { value = var.hostname }
output "username" { value = random_string.username.result }
output "password" {
  value     = random_password.password.result
  sensitive = true
}
output "database" { value = var.name }
output "uri" {
  value = format(
    "mysql://%s:%s@%s:%d/%s",
    random_string.username.result,
    random_password.password.result,
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
    var.hostname,
    var.port,
    var.name,
    random_string.username.result,
    random_password.password.result,
  )
  sensitive = true
}