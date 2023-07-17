terraform {
  required_providers {
    csbmajorengineversion = {
      source  = "cloudfoundry.org/cloud-service-broker/csbmajorengineversion"
      version = "1.0.0"
    }
  }
}

provider "csbmajorengineversion" {
  engine            = "postgres"
  access_key_id     = ""
  secret_access_key = ""
  region            = "us-west-2"
}

data "csbmajorengineversion" "major_version" {
  engine_version = "14.7"
}
