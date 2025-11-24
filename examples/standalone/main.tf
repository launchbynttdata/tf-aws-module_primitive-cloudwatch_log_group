// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "cloudwatch_logs_kms_policy" {
  statement {
    sid    = "Enable IAM User Permissions"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
    actions   = ["kms:*"]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "kms:CallerAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
  statement {
    sid    = "Allow CloudWatch Logs to use the key"
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["logs.${data.aws_region.current.name}.amazonaws.com"]
    }
    actions = [
      "kms:Encrypt*",
      "kms:Decrypt*",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
    resources = ["*"]
    condition {
      test     = "ArnEquals"
      variable = "kms:EncryptionContext:aws:logs:arn"
      values   = ["arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:*"]
    }
  }
}
module "kms_key" {
  # checkov:skip=CKV_TF_1: trusted registry source
  source                  = "terraform.registry.launch.nttdata.com/module_primitive/kms_key/aws"
  version                 = "~> 0.1"
  count                   = var.kms_key_id == null ? 1 : 0
  enable_key_rotation     = true
  deletion_window_in_days = 7
  multi_region            = true
  description             = "KMS key for Cloudwatch log groups"
  tags = {
    Name        = "terrattest-cloudwatch-log-group-kms-key"
    Environment = "terratest"
  }

}

resource "aws_kms_key_policy" "cloudwatch_logs" {
  count  = var.kms_key_id == null ? 1 : 0
  key_id = module.kms_key[0].key_id
  policy = data.aws_iam_policy_document.cloudwatch_logs_kms_policy.json
}

module "cloudwatch_log_group" {
  # regal ignore:FG_R00068
  source = "../.."

  name       = var.name
  kms_key_id = var.kms_key_id == null ? module.kms_key[0].arn : var.kms_key_id
}
