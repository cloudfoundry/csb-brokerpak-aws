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

resource "aws_s3_bucket" "b" {
  bucket = var.bucket_name
  acl    = var.acl

  versioning {
    enabled = var.enable_versioning
  }

  tags = var.labels

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_ownership_controls" "example" {
  bucket = aws_s3_bucket.b.id

  rule {
    object_ownership = var.boc_object_ownership
  }
}

resource "aws_s3_bucket_public_access_block" "bucket_public_access_block" {
  bucket = aws_s3_bucket.b.id

  block_public_acls       = var.pab_block_public_acls
  block_public_policy     = var.pab_block_public_policy
  ignore_public_acls      = var.pab_ignore_public_acls
  restrict_public_buckets = var.pab_restrict_public_buckets
}
