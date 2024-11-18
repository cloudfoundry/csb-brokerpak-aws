# Copyright 2020 Pivotal Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

output "name" { value = aws_db_instance.db_instance.db_name }
output "hostname" { value = aws_db_instance.db_instance.address }
output "username" { value = aws_db_instance.db_instance.username }
output "password" {
  value     = var.use_managed_admin_password ? "" : aws_db_instance.db_instance.password
  sensitive = true
}
output "master_secret_arn" {
  value = var.use_managed_admin_password ? aws_db_instance.db_instance.master_user_secret[0].secret_arn : ""
  sensitive = true
}
output "use_managed_admin_password" {
  value =  var.use_managed_admin_password
}

output "require_ssl" { value = var.require_ssl }
output "provider_verify_certificate" { value = var.provider_verify_certificate }
output "status" {
  value = format(
    "created db %s (id: %s) on server %s URL: https://%s.console.aws.amazon.com/rds/home?region=%s#database:id=%s;is-cluster=false",
    aws_db_instance.db_instance.db_name,
    aws_db_instance.db_instance.id,
    aws_db_instance.db_instance.address,
    var.region,
    var.region,
    aws_db_instance.db_instance.id,
  )
}

output "region" {   value = var.region  }