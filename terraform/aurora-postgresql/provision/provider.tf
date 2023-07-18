provider "aws" {
  region     = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}

provider "csbmajorengineversion" {
  region            = var.region
  engine            = local.engine
  access_key_id     = var.aws_access_key_id
  secret_access_key = var.aws_secret_access_key
}
