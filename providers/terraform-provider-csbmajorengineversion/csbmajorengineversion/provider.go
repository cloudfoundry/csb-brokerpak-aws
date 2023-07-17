// Package csbmajorengineversion is a Terraform provider designed to get the RDS major version given an engine and a reference version.
package csbmajorengineversion

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema:               ProviderSchema(),
		ConfigureContextFunc: ProviderConfigureContext,
		DataSourcesMap: map[string]*schema.Resource{
			DataResourceNameKey: DataSourceMajorEngineVersion(),
		},
	}
}

func ProviderSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		engineKey: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		awsAccessKeyIDKey: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		awsSecretAccessKeyKey: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		awsRegionKey: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
	}
}

func ProviderConfigureContext(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	tflog.Debug(ctx, "Configuring Terraform csbmajorengineversion Provider")
	engine := d.Get(engineKey).(string)
	secret := d.Get(awsSecretAccessKeyKey).(string)
	accessKey := d.Get(awsAccessKeyIDKey).(string)
	region := d.Get(awsRegionKey).(string)

	return NewEngineDescriptor(engine, accessKey, secret, region), nil
}
