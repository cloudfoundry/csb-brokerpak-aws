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

provider "csbpg" {
  host            = var.hostname
  port            = local.port
  username        = var.admin_username
  password        = var.admin_password
  database        = var.db_name
  data_owner_role = "binding_user_group"
  sslmode         = var.provider_verify_certificate ? "verify-full" : "require"
}
