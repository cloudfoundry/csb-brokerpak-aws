output "name" { value = var.db_name }
output "hostname" { value = aws_db_instance.db_instance.address }
output "username" { value = aws_db_instance.db_instance.username }
output "password" {
  value     = aws_db_instance.db_instance.password
  sensitive = true
}
output "status" { value = format("created service (id: %s) on server %s URL: https://%s.console.aws.amazon.com/rds/home?region=%s#database:id=%s;is-cluster=false",
  aws_db_instance.db_instance.id,
  aws_db_instance.db_instance.address,
  var.region,
  var.region,
aws_db_instance.db_instance.id) }
