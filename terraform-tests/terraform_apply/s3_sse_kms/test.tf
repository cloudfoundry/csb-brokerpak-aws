variable "region" { type = string }
variable "access_key" { type = string }
variable "secret_key" { type = string }

###################################################
# BEGIN: Prerequisites for running the test
###################################################

provider "aws" {
  region     = var.region
  access_key = var.access_key
  secret_key = var.secret_key
}

resource "aws_kms_key" "test_sse_kms" {
  description             = "KMS key for testing S3 SSE"
  deletion_window_in_days = 10
}

###################################################
# END: Prerequisites for running the test
###################################################



###################################################
# BEGIN: List of property<->values being tested
###################################################

###################################################
# By defining our property<->values as outputs it is very
# easy to generate a tfvars file which we can use as
# the input for the provisioning and binding steps by doing:
#
# terraform output -json provision > provision.test/terraform.tfvars.json
# terraform output -json bind > bind.test/terraform.tfvars.json
#
###################################################

########################### Provisioning properties
output "provision" {
  sensitive = true
  value = {
    region : var.region
    aws_access_key_id : sensitive(var.access_key)
    aws_secret_access_key : sensitive(var.secret_key)
    sse_default_kms_key_id : sensitive(aws_kms_key.test_sse_kms.id)

    ### Hardcoded values
    bucket_name : "fnt-csb-test-sse-kms"
    labels : { "tf_test" : "test-sse-kms" }

    acl : null
    ol_enabled : false
    require_tls : false
    enable_versioning : false
    boc_object_ownership : "BucketOwnerEnforced"

    pab_block_public_acls : false
    pab_ignore_public_acls : false
    pab_block_public_policy : false
    pab_restrict_public_buckets : false

    sse_default_algorithm : "aws:kms"
    sse_bucket_key_enabled : true
    ol_configuration_default_retention_mode : null
    ol_configuration_default_retention_days : null
    ol_configuration_default_retention_years : null
    ol_configuration_default_retention_enabled : null
  }
}

################################ Binding properties
output "bind" {
  sensitive = true
  value = {
    aws_access_key_id : sensitive(var.access_key)
    aws_secret_access_key : sensitive(var.secret_key)
    user_name : "fnt-csb-test-sse-kms"
  }
}

###################################################
# END: List of property<->values being tested
###################################################