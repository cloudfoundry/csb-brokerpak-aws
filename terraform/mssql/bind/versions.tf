terraform {

  required_providers {
    random = {
      source  = "hashicorp/random"
      version = ">= 3.3.2"
    }
    csbsqlserver = {
      source  = "cloudfoundry.org/cloud-service-broker/csbsqlserver"
      version = ">=1.0.0"
    }
  }
}