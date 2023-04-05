// Package csbdynamodbns is a Terraform provider specialised for the DynamoDB Namespace service of the AWS brokerpak
package csbdynamodbns

import (
	"context"
	"net/url"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	awsRegionKey         = "region"
	dynamoDBPrefixKey    = "prefix"
	customEndpointURLKey = "custom_endpoint_url"
)

var identifierRegexp = regexp.MustCompile(`^[\w_.-]{1,64}$`)

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
			"csbdynamodbns_instance": ResourceDynamoDBNSInstance(),
		},
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var (
		diags             diag.Diagnostics
		region            string
		prefix            string
		customEndpointURL string
	)

	for _, f := range []func() diag.Diagnostics{
		func() (dg diag.Diagnostics) {
			region, dg = getIdentifier(d, awsRegionKey)
			return
		},
		func() (dg diag.Diagnostics) {
			prefix, dg = getIdentifier(d, dynamoDBPrefixKey)
			return
		},
		func() diag.Diagnostics {
			if customURL, ok := d.GetOk(customEndpointURLKey); ok {
				uri, err := url.ParseRequestURI(customURL.(string))
				if err != nil {
					return diag.FromErr(err)
				}
				customEndpointURL = uri.String()
			}
			return nil
		},
	} {
		if dg := f(); dg != nil {
			return nil, dg
		}
	}

	var settings = &dynamoDBNamespaceSettings{
		region:            region,
		prefix:            prefix,
		customEndpointURL: customEndpointURL,
	}

	return settings, diags
}

func getIdentifier(d *schema.ResourceData, key string) (string, diag.Diagnostics) {
	// We rely on Terraform to supply the correct types, and it's ok panic if this contract is broken
	s := d.Get(key).(string)
	if !identifierRegexp.MatchString(s) {
		return "", diag.Errorf("invalid value %q for identifier %q, validation expression is: %s", s, key, identifierRegexp.String())
	}

	return s, nil
}
