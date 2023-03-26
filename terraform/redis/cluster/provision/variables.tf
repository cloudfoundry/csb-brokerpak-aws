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

variable "cache_size" { type = number }
variable "redis_version" { type = string }
variable "instance_name" { type = string }
variable "labels" { type = map(any) }
variable "aws_vpc_id" { type = string }
variable "node_type" { type = string }
variable "node_count" { type = number }
variable "elasticache_subnet_group" { type = string }
variable "elasticache_vpc_security_group_ids" { type = string }
variable "at_rest_encryption_enabled" { type = bool }
variable "automatic_failover_enabled" { type = bool }
variable "kms_key_id" { type = string }
variable "maintenance_day" { type = string }
variable "maintenance_start_hour" { type = string }
variable "maintenance_start_min" { type = string }
variable "maintenance_end_hour" { type = string }
variable "maintenance_end_min" { type = string }
variable "data_tiering_enabled" { type = bool }
variable "multi_az_enabled" { type = bool }
variable "backup_retention_limit" { type = number }
variable "final_backup_identifier" { type = string }
variable "backup_name" { type = string }
variable "backup_start_hour" { type = string }
variable "backup_start_min" { type = string }
variable "backup_end_hour" { type = string }
variable "backup_end_min" { type = string }
variable "parameter_group_name" { type = string }
variable "preferred_azs" { type = list(string) }