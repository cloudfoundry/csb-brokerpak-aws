terraform {
  required_providers {
    aws = {
      source  = "registry.terraform.io/hashicorp/aws"
      version = ">= 5.0"
    }
    random = {
      source  = "registry.terraform.io/hashicorp/random"
      version = ">= 3.3.2"
    }
  }
}