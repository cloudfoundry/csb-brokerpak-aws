# Run "make init" to perform "terraform init"
# The easiest way to get a DynamoDB is: docker run -p 8000:8000 -t amazon/dynamodb-local

terraform {
  required_providers {
    csbmajorengineversion = {
      source  = "cloudfoundry.org/cloud-service-broker/csbmajorengineversion"
      version = "1.0.0"
    }
  }
}

provider "csbmajorengineversion" {
  engine = "postgres"
}

data "csbmajorengineversion" "major_version" {
  engine_version     = "14.7"
}
