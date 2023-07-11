package csbmajorengineversion

import (
	"context"
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

	if err != nil {
		return diag.FromErr(err)
	}

	awsClient := rds.NewFromConfig(cfg)
	engine := settings.engine
	engineVersion := data.Get(EngineVersionKey).(string)

	output, err := awsClient.DescribeDBEngineVersions(context.Background(), &rds.DescribeDBEngineVersionsInput{
		Engine:        &engine,
		EngineVersion: &engineVersion,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if len(output.DBEngineVersions) == 0 {
		return diag.FromErr(err)
	}
	data.SetId("version")
	data.Set(MajorVersionKey, output.DBEngineVersions[0].MajorEngineVersion)
	return nil

}
