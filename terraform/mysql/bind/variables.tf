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

variable "db_name" { type = string }
variable "hostname" { type = string }
variable "admin_username" { type = string }
variable "admin_password" { type = string }
variable "require_ssl" { type = bool }

locals {
  port = 3306

  # The following values are allowed:
  # "DISABLED" - Establish unencrypted connections;
  # "PREFERRED" - Establish encrypted connections if the server enabled them, otherwise fall back to unencrypted connections;
  # "REQUIRED" - Establish secure connections if the server enabled them, fail otherwise;
  # "VERIFY_CA" - Like "REQUIRED" but additionally verify the server TLS certificate against the configured Certificate Authority (CA) certificates;
  # "VERIFY_IDENTITY" - Like "VERIFY_CA", but additionally verify that the server certificate matches the host to which the connection is attempted.
  sslMode = var.require_ssl ? "REQUIRED" : "PREFERRED"
}