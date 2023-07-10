// Package csbmajorengineversion is a Terraform provider specialised for the DynamoDB Namespace service of the AWS brokerpak
package csbmajorengineversion

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	engineKey             = "engine"
	awsAccessKeyIDKey     = "access_key_id"
	awsSecretAccessKeyKey = "secret_access_key"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			engineKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			awsAccessKeyIDKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			awsSecretAccessKeyKey: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		ConfigureContextFunc: providerConfigure,
		DataSourcesMap: map[string]*schema.Resource{
			"csbmajorengineversion": DataSourceMajorEngineVersion(),
		},
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var (
		diags                     diag.Diagnostics
		engine, secret, accessKey string
	)

	engine = d.Get(engineKey).(string)
	secret = d.Get(awsSecretAccessKeyKey).(string)
	accessKey = d.Get(awsAccessKeyIDKey).(string)

	var settings = &RDSSettings{
		engine:             engine,
		AWSAccessKeyID:     accessKey,
		AWSSecretAccessKey: secret,
	}

	return settings, diags
}
