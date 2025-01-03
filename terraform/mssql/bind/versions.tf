terraform {

  required_providers {
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = "~> 3"
    }
    csbsqlserver = {
      source  = "cloudfoundry.org/cloud-service-broker/csbsqlserver"
      version = ">=1.0.0"
    }
    aws = {
      source  = "registry.terraform.io/hashicorp/aws"
      version = ">= 5.0"
    }
  }
}