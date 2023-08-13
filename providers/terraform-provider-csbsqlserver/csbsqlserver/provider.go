// Package csbsqlserver is a niche Terraform provider for Microsoft SQL Server
package csbsqlserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-csbsqlserver/connector"
)

const (
	serverKey           = "server"
	portKey             = "port"
	databaseKey         = "database"
	providerUsernameKey = "username"
	providerPasswordKey = "password"
	encryptKey          = "encrypt"
	iaasKey             = "iaas"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			serverKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			portKey: {
				Type:     schema.TypeInt,
				Required: true,
			},
			databaseKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			providerUsernameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			providerPasswordKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			encryptKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
			iaasKey: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      connector.AWS,
				ValidateFunc: validation.StringInSlice([]string{connector.AWS, connector.Azure}, false),
			},
		},
		ConfigureContextFunc: configure,
		ResourcesMap: map[string]*schema.Resource{
			"csbsqlserver_binding": bindingResource(),
		},
	}
}

func configure(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var (
		server   string
		port     int
		username string
		password string
		database string
		encrypt  string
		iaas     string
	)

	for _, f := range []func() diag.Diagnostics{
		func() (diags diag.Diagnostics) {
			server, diags = getURL(d, serverKey)
			return
		},
		func() (diags diag.Diagnostics) {
			port, diags = getPort(d, portKey)
			return
		},
		func() (diags diag.Diagnostics) {
			username, diags = getServerIdentifier(d, providerUsernameKey)
			return
		},
		func() (diags diag.Diagnostics) {
			password, diags = getServerPassword(d, providerPasswordKey)
			return
		},
		func() (diags diag.Diagnostics) {
			database, diags = getServerIdentifier(d, databaseKey)
			return
		},
		func() (diags diag.Diagnostics) {
			encrypt, diags = getEncrypt(d, encryptKey)
			return
		},
		func() (diags diag.Diagnostics) {
			iaas, diags = getServerIdentifier(d, iaasKey)
			return
		},
	} {
		if d := f(); d != nil {
			return nil, d
		}
	}

	var e = connector.NewEncoder(
		server,
		username,
		password,
		database,
		encrypt,
		iaas,
		port,
	)

	return connector.New(server, port, username, password, database, encrypt, e), nil
}
