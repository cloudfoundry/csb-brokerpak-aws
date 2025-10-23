output "name" { value = aws_rds_cluster.cluster.database_name }
output "hostname" { value = aws_rds_cluster.cluster.endpoint }
output "reader_hostname" { value = aws_rds_cluster.cluster.reader_endpoint }
output "port" { value = var.port }
output "username" { value = aws_rds_cluster.cluster.master_username }
output "password" {
  value     = var.use_managed_admin_password ? "" : aws_rds_cluster.cluster.master_password
  sensitive = true
}
output "managed_admin_credentials_arn" {
  # Using join and master_user_secret.*.secret_arn is a workaround to make sure that the value of the secret ARN is evaluated after the apply.
  # There is currently a bug which results in no value evaluated if using the usual syntax aws_rds_cluster.cluster.master_user_secret[0].secret_arn
  # when updating from a password db to a managed secret db. See: https://github.com/hashicorp/terraform-provider-aws/issues/34094
  # Note that aws_rds_cluster.cluster.master_user_secret always returns max one item.
  value     = var.use_managed_admin_password ? join("", aws_rds_cluster.cluster.master_user_secret.*.secret_arn) : ""
  sensitive = true
}
output "use_managed_admin_password" {
  value = var.use_managed_admin_password
}
output "region" {
  value = var.region
}
output "status" {
  value = format(
    "created db %s (id: %s) on server %s",
    aws_rds_cluster.cluster.database_name,
    aws_rds_cluster.cluster.cluster_identifier,
    aws_rds_cluster.cluster.endpoint,
  )
}
