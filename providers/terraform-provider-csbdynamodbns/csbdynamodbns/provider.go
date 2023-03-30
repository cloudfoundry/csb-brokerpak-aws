// Package csbdynamodbns is a Terraform provider specialised for the DynamoDB Namespace service of the AWS brokerpak
package csbdynamodbns

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	awsRegionKey         = "region"
	dynamoDBPrefixKey    = "prefix"
	customEndpointURLKey = "custom_endpoint_url"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			awsRegionKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			dynamoDBPrefixKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			customEndpointURLKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ConfigureContextFunc: providerConfigure,
		ResourcesMap: map[string]*schema.Resource{
			"csbdynamodbns_instance": resourceDynamoDBNSInstance(),
		},
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var settings = &dynamoDBNamespaceSettings{
		region: d.Get(awsRegionKey).(string),
		prefix: d.Get(dynamoDBPrefixKey).(string),
	}

	if customEndpointURL, ok := d.GetOk(customEndpointURLKey); ok {
		settings.customEndpointURL = customEndpointURL.(string)
	}

	return settings, diags
}
