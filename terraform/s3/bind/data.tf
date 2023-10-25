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

locals {
  user_policy_with_or_without_encryption = try(data.aws_iam_policy_document.user_policy_sse[0], data.aws_iam_policy_document.user_policy)
}


data "aws_iam_policy_document" "user_policy" {
  statement {
    sid = "bucketAccess"
    actions = [
      "s3:ListBucket",
      "s3:ListBucketVersions",
      "s3:ListBucketMultipartUploads",
      "s3:GetBucketCORS",
      "s3:PutBucketCORS",
      "s3:GetBucketVersioning",
      "s3:PutBucketVersioning",
      "s3:GetBucketRequestPayment",
      "s3:PutBucketRequestPayment",
      "s3:GetBucketLocation",
      "s3:GetBucketNotification",
      "s3:PutBucketNotification",
      "s3:GetBucketLogging",
      "s3:PutBucketLogging",
      "s3:GetBucketTagging",
      "s3:PutBucketTagging",
      "s3:GetBucketWebsite",
      "s3:PutBucketWebsite",
      "s3:DeleteBucketWebsite",
      "s3:GetLifecycleConfiguration",
      "s3:PutLifecycleConfiguration",
      "s3:PutReplicationConfiguration",
      "s3:GetReplicationConfiguration",
      "s3:DeleteReplicationConfiguration",
    ]
    resources = [
      var.arn
    ]
  }

  statement {
    sid = "bucketContentAccess"
    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion",
      "s3:PutObject",
      "s3:GetObjectAcl",
      "s3:GetObjectVersionAcl",
      "s3:PutObjectAcl",
      "s3:PutObjectVersionAcl",
      "s3:DeleteObject",
      "s3:DeleteObjectVersion",
      "s3:ListMultipartUploadParts",
      "s3:AbortMultipartUpload",
      "s3:GetObjectTorrent",
      "s3:GetObjectVersionTorrent",
      "s3:RestoreObject",
    ]
    resources = [
      format("%s/*", var.arn)
    ]
  }
}

data "aws_kms_key" "customer_provided_key" {
  count  = try(length(var.sse_default_kms_key_id), 0) == 0 ? 0 : 1
  key_id = var.sse_default_kms_key_id
}

data "aws_iam_policy_document" "user_policy_sse" {
  count = try(length(var.sse_default_kms_key_id), 0) == 0 ? 0 : 1

  source_policy_documents = [data.aws_iam_policy_document.user_policy.json]

  statement {
    sid = "kmsperms"
    actions = [
      "kms:Decrypt",
      "kms:Encrypt",
      "kms:GenerateDataKey",
    ]
    resources = [data.aws_kms_key.customer_provided_key[0].arn]
  }
}
