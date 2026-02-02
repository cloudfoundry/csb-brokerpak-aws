provider "aws" {
  region = var.region
}

provider "csbdynamodbns" {
  region = var.region
  prefix = var.prefix
}
