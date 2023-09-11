output "name" { value = var.db_name }
output "hostname" { value = aws_db_instance.db_instance.address }
output "username" { value = aws_db_instance.db_instance.username }
output "password" {
  value     = aws_db_instance.db_instance.password
  sensitive = true
}
output "require_ssl" { value = var.require_ssl }
output "status" {
  value = format(
    "created service (id: %s) on server %s - region %s",
    aws_db_instance.db_instance.id,
    aws_db_instance.db_instance.address,
    var.region,
  )
}
