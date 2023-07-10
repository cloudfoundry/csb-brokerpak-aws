package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-majorengineversion/csbmajorengineversion"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbmajorengineversion.Provider,
	})
}
