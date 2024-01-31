terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.3.2"
    }
    csbmajorengineversion = {
      source  = "cloudfoundry.org/cloud-service-broker/csbmajorengineversion"
      version = "1.0.0"
    }
  }
}