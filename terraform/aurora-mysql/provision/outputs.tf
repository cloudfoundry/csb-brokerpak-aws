output "name" { value = aws_rds_cluster.cluster.database_name }
output "hostname" { value = aws_rds_cluster.cluster.endpoint }
output "username" { value = aws_rds_cluster.cluster.master_username }
output "port" { value = aws_rds_cluster.cluster.port }
output "password" {
  value     = aws_rds_cluster.cluster.master_password
  sensitive = true
}
output "status" {
  value = format(
    "created db %s (id: %s) on server %s",
    aws_rds_cluster.cluster.database_name,
    aws_rds_cluster.cluster.cluster_identifier,
    aws_rds_cluster.cluster.endpoint,
  )
}