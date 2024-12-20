terraform {
  required_providers {
    csbmysql = {
      source  = "cloudfoundry.org/cloud-service-broker/csbmysql"
      version = ">= 1.0.0"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = "~> 3"
    }
    aws = {
      source  = "registry.terraform.io/hashicorp/aws"
      version = ">= 5.0"
    }
  }
}