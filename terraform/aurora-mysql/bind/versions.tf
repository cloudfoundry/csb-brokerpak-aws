terraform {
  required_providers {
    csbmysql = {
      source  = "cloud-service-broker/csbmysql"
      version = ">= 1.0.0"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = ">= 3.3.2"
    }
  }
}