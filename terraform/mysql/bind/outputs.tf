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

output "username" { value = csbmysql_binding_user.new_user.username }
output "password" {
  value     = csbmysql_binding_user.new_user.password
  sensitive = true
}
output "uri" {
  value = format(
    "mysql://%s:%s@%s:%d/%s",
    csbmysql_binding_user.new_user.username,
    csbmysql_binding_user.new_user.password,
    var.hostname,
    local.port,
    var.db_name,
  )
  sensitive = true
}
output "port" { value = local.port }
output "jdbcUrl" {
  value = format(
    "jdbc:mysql://%s:%d/%s?user=%s\u0026password=%s\u0026useSSL=true",
    var.hostname,
    local.port,
    var.db_name,
    csbmysql_binding_user.new_user.username,
    csbmysql_binding_user.new_user.password,
  )
  sensitive = true
}