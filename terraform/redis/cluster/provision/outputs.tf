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

output "name" { value = aws_elasticache_replication_group.redis.id }
output "host" { value = aws_elasticache_replication_group.redis.primary_endpoint_address }
output "password" {
  value     = random_password.auth_token.result
  sensitive = true
}
output "tls_port" { value = local.port }
output "status" {
  value = format("created cache %s (id: %s)", aws_elasticache_replication_group.redis.primary_endpoint_address, aws_elasticache_replication_group.redis.id)
}
output "reader_endpoint" { value = aws_elasticache_replication_group.redis.reader_endpoint_address }
