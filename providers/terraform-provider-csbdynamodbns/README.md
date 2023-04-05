# terraform-provider-dynamodbns

This is a highly specialised Terraform provider designed to be used exclusively with the [Cloud Service Broker](https://github.com/cloudfoundry/cloud-service-broker) ("CSB") in the `dynamodb-namespace` service of the AWS brokerpak.

The Cloud Service Broker complies with the [OSBAPI](https://www.openservicebrokerapi.org) model, according to which, every time a service instance is deleted, all of its associated data should also be deleted. The `dynamodb-namespace` does not create any database objects itself, but rather provides the service user with credentials for that purpose. The purpose of the `terraform-provider-dynamodbns`, therefore, is to clean up any DynamoDB tables that could have been left over before the service instance is deleted.

## Notes

The user account supplied to the resource must have `ListTables` permissions on all tables in addition to `DeleteTable` permission for tables with a given prefix.