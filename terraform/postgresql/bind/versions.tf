terraform {
  required_providers {
    csbpg = {
      source  = "cloud-service-broker/csbpg"
      version = ">= 1.0.1"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.3.2"
    }
  }
}