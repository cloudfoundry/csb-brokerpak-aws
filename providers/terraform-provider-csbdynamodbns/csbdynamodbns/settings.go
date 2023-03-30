package csbdynamodbns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type dynamoDBNamespaceSettings struct {
	region            string
	prefix            string
	customEndpointURL string
}

func (d *dynamoDBNamespaceSettings) GetClient(ctx context.Context, keyID, secretKey string) (*dynamodb.Client, error) {
	optFns := []func(options *config.LoadOptions) error{
		config.WithRegion(d.region),
	}
	if d.customEndpointURL == "" {
		optFns = append(optFns,
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
	} else {
		optFns = append(optFns,
			config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(
					func(service, region string, options ...interface{}) (aws.Endpoint, error) {
						return aws.Endpoint{URL: d.customEndpointURL}, nil
					},
				),
			),
		)
	}
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}
