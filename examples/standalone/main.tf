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

module "resource_names" {
  # checkov:skip=CKV_TF_1: trusted registry source
  source  = "terraform.registry.launch.nttdata.com/module_library/resource_name/launch"
  version = "~> 2.0"

  logical_product_family  = "cloudwatch"
  logical_product_service = "loggroup"
  region                  = "us-east-2"
  class_env               = "terratest"
  instance_env            = "000"
  cloud_resource_type     = "loggroup"
  instance_resource       = "000"
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

module "kms_key_policy" {
  # checkov:skip=CKV_TF_1: trusted registry source
  source  = "terraform.registry.launch.nttdata.com/module_primitive/kms_key_policy/aws"
  version = "~> 0.1"
  count   = var.kms_key_id == null ? 1 : 0
  key_id  = module.kms_key[0].key_id
  policy  = data.aws_iam_policy_document.cloudwatch_logs_kms_policy
}

module "cloudwatch_log_group" {
  source = "../.."

  name       = var.name != null ? var.name : module.resource_names.minimal_random_suffix
  kms_key_id = var.kms_key_id == null ? module.kms_key[0].arn : var.kms_key_id
}
