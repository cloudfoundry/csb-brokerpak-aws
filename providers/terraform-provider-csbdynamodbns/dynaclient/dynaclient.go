// Package dynaclient is a helper to create a DynamoDB client used both in test and implementation
package dynaclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func New(ctx context.Context, region, keyID, secretKey, customEndpointURL string) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
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
	if customEndpointURL != "" {
		opts = append(opts, dynamodb.WithEndpointResolverV2(endpointResolverV2{endpoint: customEndpointURL}))
	}

	return dynamodb.NewFromConfig(cfg, opts...), nil
}
