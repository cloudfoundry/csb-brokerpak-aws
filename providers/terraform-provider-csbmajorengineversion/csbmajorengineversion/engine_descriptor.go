package csbmajorengineversion

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type engineDescriptor struct {
	engine          string
	accessKeyID     string
	secretAccessKey string
	region          string
}

func NewEngineDescriptor(engine, accessKeyID, secretAccessKey, region string) *engineDescriptor {
	return &engineDescriptor{engine: engine, accessKeyID: accessKeyID, secretAccessKey: secretAccessKey, region: region}
}

func (e *engineDescriptor) Describe(ctx context.Context, engineVersion string) (string, error) {
	credentialsCache := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(e.accessKeyID, e.secretAccessKey, ""))

	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credentialsCache))
	if err != nil {
		return "", err
	}

	tflog.Debug(ctx, "Retrieving AWS DB engine versions", map[string]any{
		"engine":         e.engine,
		"engine_version": engineVersion,
	})
	awsClient := rds.NewFromConfig(cfg)
	params := &rds.DescribeDBEngineVersionsInput{
		Engine:        aws.String(e.engine),
		EngineVersion: aws.String(engineVersion),
		IncludeAll:    aws.Bool(true), // If false, Postgres version 14.2 does not return any output because it is no longer listed in the AWS console
	}
	optFns := func(options *rds.Options) { options.Region = e.region }
	output, err := awsClient.DescribeDBEngineVersions(ctx, params, optFns)
	if err != nil {
		return "", err
	}

	if len(output.DBEngineVersions) == 0 {
		return "", fmt.Errorf(
			"invalid parameter combination. API does not return any db engine version - engine %s - engine version %s",
			e.engine,
			engineVersion,
		)
	}

	return aws.ToString(output.DBEngineVersions[0].MajorEngineVersion), nil
}
