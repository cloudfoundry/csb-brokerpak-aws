terraform {
#  required_version = ">= 1.4.4"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.3.2"
    }
    csbsqlserver = {
      source  = "cloudfoundry.org/cloud-service-broker/csbsqlserver"
      version = "1.0.0"
    }
  }
}