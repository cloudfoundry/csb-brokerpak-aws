terraform {
  required_providers {
    mysql = {
      source  = "hashicorp/mysql"
      version = ">= 1.9"
    }
    csbmysql = {
      source  = "cloud-service-broker/csbmysql"
      version = ">= 0.0.3"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.3.2"
    }
  }
}