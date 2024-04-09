terraform {
  required_providers {
    aws = {
      source  = "registry.terraform.io/hashicorp/aws"
      version = ">= 5.0"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = ">= 3.5.1"
    }
    csbmajorengineversion = {
      source  = "cloudfoundry.org/cloud-service-broker/csbmajorengineversion"
      version = "1.0.0"
    }
  }
}