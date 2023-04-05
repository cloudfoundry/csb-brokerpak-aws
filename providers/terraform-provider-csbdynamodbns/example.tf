# Run "make init" to perform "terraform init"
# The easiest way to get a DynamoDB is: docker run -p 8000:8000 -t amazon/dynamodb-local

terraform {
  required_providers {
    csbdynamodbns = {
      source  = "cloudfoundry.org/cloud-service-broker/csbdynamodbns"
      version = "1.0.0"
    }
  }
}

provider "csbdynabodbns" {
  region = "us-west-2"
  prefix = "csb-46d6f6fb-c746-4488-8ed9-bc05bff03eb8"
}

resource "csbdynamodbns_instance" "service_instance" {
  access_key_id     = "FAKE-access-key-id"
  secret_access_key = "FAKE-secret-access-key"
}
