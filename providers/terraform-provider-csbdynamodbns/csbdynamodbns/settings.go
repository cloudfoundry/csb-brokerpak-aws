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

//counterfeiter:generate -header csbdynamodbnsfakes/header.txt . DynamoDBConfig
type DynamoDBConfig interface {
	GetClient(ctx context.Context, keyID, secretKey string) (DynamoDBClient, error)
	GetPrefix() string
}

var _ DynamoDBConfig = &dynamoDBNamespaceSettings{}

func (d *dynamoDBNamespaceSettings) GetPrefix() string {
	return d.prefix
}

func (d *dynamoDBNamespaceSettings) GetClient(ctx context.Context, keyID, secretKey string) (DynamoDBClient, error) {
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
			config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     "dummy",
					SecretAccessKey: "dummy",
					SessionToken:    "dummy",
					Source:          "Hard-coded credentials; the values are irrelevant",
				},
			}),
		)
	}
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}
