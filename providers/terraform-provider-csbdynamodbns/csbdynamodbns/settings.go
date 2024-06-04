package csbdynamodbns

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

//counterfeiter:generate -header csbdynamodbnsfakes/header.txt . DynamoDBConfig
type DynamoDBConfig interface {
	GetClient(ctx context.Context, keyID, secretKey string) (DynamoDBClient, error)
	GetPrefix() string
}

type endpointResolverV2Func func(ctx context.Context, params dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error)

// Fail fast if the interface is not implemented
var _ dynamodb.EndpointResolverV2 = endpointResolverV2Func(nil)

func (e endpointResolverV2Func) ResolveEndpoint(ctx context.Context, params dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	return e(ctx, params)
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
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(d.region),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(
				credentials.NewStaticCredentialsProvider(
					keyID,
					secretKey,
					"",
				),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// For testing we use a custom endpoint
	var opts []func(*dynamodb.Options)
	if d.customEndpointURL != "" {
		u, err := url.Parse(d.customEndpointURL)
		if err != nil {
			return nil, err
		}

		opts = append(opts, dynamodb.WithEndpointResolverV2(endpointResolverV2Func(func(ctx context.Context, params dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
			return smithyendpoints.Endpoint{URI: *u}, nil
		})))
	}

	return dynamodb.NewFromConfig(cfg, opts...), nil
}
