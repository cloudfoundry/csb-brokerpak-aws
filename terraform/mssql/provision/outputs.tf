output "name" { value = var.db_name }
output "hostname" { value = aws_db_instance.db_instance.address }
output "username" { value = aws_db_instance.db_instance.username }
output "password" {
  value     = var.use_managed_admin_password ? "" : aws_db_instance.db_instance.password
  sensitive = true
}
output "managed_admin_credentials_arn" {
  # Using join and master_user_secret.*.secret_arn is a workaround to make sure that the value of the secret ARN is evaluated after the apply.
  # There is currently a bug which results in no value evaluated if using the usual syntax aws_db_instance.db_instance.master_user_secret[0].secret_arn
  # when updating from a password db to a managed secret db. See: https://github.com/hashicorp/terraform-provider-aws/issues/34094
  # Note that aws_db_instance.db_instance.master_user_secret always returns max one item.
  value     = var.use_managed_admin_password ? join("", aws_db_instance.db_instance.master_user_secret.*.secret_arn) : ""
  sensitive = true
}
output "use_managed_admin_password" {
  value = var.use_managed_admin_password
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
output "region" { value = var.region }