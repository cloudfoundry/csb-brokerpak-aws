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

output "arn" { value = aws_s3_bucket.b.arn }
output "bucket_domain_name" { value = aws_s3_bucket.b.bucket_domain_name }
output "region" { value = aws_s3_bucket.b.region }
output "bucket_name" { value = aws_s3_bucket.b.bucket }
output "sse_all_kms_key_ids" { value = local.sse_all_kms_key_ids }
