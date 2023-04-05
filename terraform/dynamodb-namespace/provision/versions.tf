terraform {
  required_providers {
    csbdynamodbns = {
      source  = "cloudfoundry.org/cloud-service-broker/csbdynamodbns"
      version = "1.0.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
  }
}