package csbdynamodbns

import (
	"context"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-dynamodbns/dynaclient"
)

//counterfeiter:generate -header csbdynamodbnsfakes/header.txt . DynamoDBConfig
type DynamoDBConfig interface {
	GetClient(ctx context.Context, keyID, secretKey string) (DynamoDBClient, error)
	GetPrefix() string
}

type dynamoDBNamespaceSettings struct {
	region            string
	prefix            string
	customEndpointURL string
}

// Fail fast if the interface is not implemented
var _ DynamoDBConfig = &dynamoDBNamespaceSettings{}

func (d *dynamoDBNamespaceSettings) GetPrefix() string {
	return d.prefix
}

func (d *dynamoDBNamespaceSettings) GetClient(ctx context.Context, keyID, secretKey string) (DynamoDBClient, error) {
	return dynaclient.New(ctx, d.region, keyID, secretKey, d.customEndpointURL)
}
