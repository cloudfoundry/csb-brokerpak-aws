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
  bucket              = var.bucket_name
  object_lock_enabled = var.ol_enabled

  tags = var.labels

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  count  = (var.acl != null) ? 1 : 0
  bucket = aws_s3_bucket.b.id
  acl    = var.acl

  depends_on = [
    aws_s3_bucket_ownership_controls.bucket_ownership_controls
  ]
}

resource "aws_s3_bucket_versioning" "bucket_versioning" {
  bucket = aws_s3_bucket.b.id
  versioning_configuration {
    status = local.is_versioning_enabled ? "Enabled" : "Disabled"
  }
}

resource "aws_s3_bucket_ownership_controls" "bucket_ownership_controls" {
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

resource "aws_s3_bucket_server_side_encryption_configuration" "server_side_encryption_configuration" {
  count  = (var.sse_default_algorithm != null || var.sse_bucket_key_enabled != false) ? 1 : 0
  bucket = aws_s3_bucket.b.bucket

  rule {
    dynamic "apply_server_side_encryption_by_default" {
      for_each = var.sse_default_algorithm[*]
      content {
        kms_master_key_id = var.sse_default_kms_key_id
        sse_algorithm     = var.sse_default_algorithm
      }
    }

    bucket_key_enabled = var.sse_bucket_key_enabled
  }
}

# aws_s3_bucket_object_lock_configuration.object_lock_enabled only admits the "Enabled" value. To be able to disable
# the resource, we have to remove it.
resource "aws_s3_bucket_object_lock_configuration" "bucket_object_lock_configuration" {
  count               = local.ol_configuration_is_enabled ? 1 : 0
  bucket              = aws_s3_bucket.b.id
  object_lock_enabled = "Enabled"

  rule {
    default_retention {
      mode  = var.ol_configuration_default_retention_mode
      days  = var.ol_configuration_default_retention_days
      years = var.ol_configuration_default_retention_years
    }
  }
}

resource "aws_s3_bucket_policy" "bucket_restrict_to_tls_requests_only_policy" {
  count = var.restrict_to_tls_requests_only ? 1 : 0
  bucket = aws_s3_bucket.b.id
  policy = jsonencode({
    "Version": "2012-10-17",
    "Statement": [{
      "Sid": "RestrictToTLSRequestsOnly",
      "Action": "s3:*",
      "Effect": "Deny",
      "Resource": [
        "${aws_s3_bucket.b.arn}",
        "${aws_s3_bucket.b.arn}/*"
      ],
      "Condition": {
        "Bool": {
          "aws:SecureTransport": "false"
        }
      },
      "Principal": "*"
    }]
  })
}