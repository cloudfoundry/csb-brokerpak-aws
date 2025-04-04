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

data "aws_secretsmanager_secret_version" "secret-version" {
  count     = var.use_managed_admin_password ? 1 : 0
  secret_id = var.managed_admin_credentials_arn
}

locals {
  managed_admin_creds    = var.use_managed_admin_password ? jsondecode(data.aws_secretsmanager_secret_version.secret-version[0].secret_string) : {}
  managed_admin_password = var.use_managed_admin_password ? local.managed_admin_creds.password : ""
}