package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-csbsqlserver/csbsqlserver"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbsqlserver.Provider,
	})
}
