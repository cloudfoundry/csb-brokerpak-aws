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
  binding_role = try(aws_iam_role.new_role[0], data.aws_iam_role.provided[0])
}

resource "random_password" "source_identity" {
  length  = 16
  special = false
}

data "aws_iam_role" "provided" {
  count = length(var.role_name) == 0 ? 0 : 1

  name = var.role_name
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    sid = "assumeRole"
    actions = [
      "sts:AssumeRole",
      "sts:TagSession",
    ]
    principals {
      type        = "AWS"
      identifiers = [var.iam_arn]
    }
    condition {
      test     = "StringEquals"
      values   = [random_password.source_identity.result]
      variable = "aws:RequestTag/current-binding"
    }
  }
}

data "aws_iam_policy_document" "dynamo_access" {
  statement {
    sid = "dynamoAccess"
    actions = [
      "dynamodb:*",
    ]
    resources = [
      var.dynamodb_table_arn
    ]
    condition {
      test     = "StringEquals"
      values   = [random_password.source_identity.result]
      variable = "aws:PrincipalTag/current-binding"
    }
  }
}
