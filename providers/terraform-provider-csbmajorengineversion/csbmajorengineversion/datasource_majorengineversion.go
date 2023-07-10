package csbmajorengineversion

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	EngineVersionKey = "engine_version"
	MajorVersionKey  = "major_version"
)

func DataSourceMajorEngineVersion() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			EngineVersionKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			MajorVersionKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		ReadContext: ResourceMajorEngineVersionRead,
		Description: "Returns major version value",
	}
}

func ResourceMajorEngineVersionRead(ctx context.Context, data *schema.ResourceData, providerConfig any) diag.Diagnostics {
	settings := providerConfig.(*RDSSettings)

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(settings.AWSAccessKeyID, settings.AWSSecretAccessKey, ""))),
	)
	data.SetId("version")
	if err != nil {
		data.Set(MajorVersionKey, fmt.Sprintf("Error returned from API %s", err))
		return nil
	}

	awsClient := rds.NewFromConfig(cfg)
	engine := settings.engine
	engineVersion := data.Get(EngineVersionKey).(string)

	output, err := awsClient.DescribeDBEngineVersions(context.Background(), &rds.DescribeDBEngineVersionsInput{
		Engine:        &engine,
		EngineVersion: &engineVersion,
	})
	if err != nil {
		data.Set(MajorVersionKey, fmt.Sprintf("Error returned from API %s. Engine %s; engine version %s", err, engine, engineVersion))
		return nil
	}

	if len(output.DBEngineVersions) == 0 {
		data.Set("error", fmt.Sprintf("No engine versions returned from API. Engine %s; engine version %s", engine, engineVersion))
		data.Set(MajorVersionKey, "No data returned from API")
		return nil
	}

	data.Set(MajorVersionKey, output.DBEngineVersions[0].MajorEngineVersion)
	return nil

}
