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

variable "bucket_name" { type = string }
variable "acl" { type = string }
variable "labels" { type = map(any) }
variable "enable_versioning" { type = bool }
variable "ol_enabled" { type = bool }
variable "boc_object_ownership" { type = string }

# Resource aws_s3_bucket_public_access_block
variable "pab_block_public_acls" { type = bool }
variable "pab_block_public_policy" { type = bool }
variable "pab_ignore_public_acls" { type = bool }
variable "pab_restrict_public_buckets" { type = bool }

# Resource aws_s3_bucket_server_side_encryption_configuration
variable "sse_default_kms_key_id" { type = string }
variable "sse_default_algorithm" { type = string }
variable "sse_bucket_key_enabled" { type = string }

# Resource aws_s3_bucket_object_lock_configuration
variable "ol_configuration_default_retention_enabled" { type = bool }
variable "ol_configuration_default_retention_mode" { type = string }
variable "ol_configuration_default_retention_days" { type = number }
variable "ol_configuration_default_retention_years" { type = number }

