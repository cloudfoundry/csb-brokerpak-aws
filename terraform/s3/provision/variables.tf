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

variable "types" {
  default = {}
  type = object({
    region                                     = optional(string)
    bucket_name                                = optional(string)
    acl                                        = optional(string)
    labels                                     = optional(map(any))
    enable_versioning                          = optional(bool)
    ol_enabled                                 = optional(bool)
    boc_object_ownership                       = optional(string)
    pab_block_public_acls                      = optional(bool)
    pab_block_public_policy                    = optional(bool)
    pab_ignore_public_acls                     = optional(bool)
    pab_restrict_public_buckets                = optional(bool)
    sse_default_kms_key_id                     = optional(string)
    sse_extra_kms_key_ids                      = optional(string)
    sse_default_algorithm                      = optional(string)
    sse_bucket_key_enabled                     = optional(bool)
    aws_s3_bucket_object_lock_configuration    = optional(bool)
    ol_configuration_default_retention_enabled = optional(bool)
    ol_configuration_default_retention_mode    = optional(string)
    ol_configuration_default_retention_days    = optional(number)
    ol_configuration_default_retention_years   = optional(number)
    require_tls                                = optional(bool)
  })
}

variable "inputs" {
  type = any
  default = {}
}

