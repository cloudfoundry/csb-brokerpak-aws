package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.gwd.broadcom.net/TNZ/csb-brokerpak-aws/terraform-provider-majorengineversion/csbmajorengineversion"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		Debug:        debug,
		ProviderFunc: csbmajorengineversion.Provider,
	})
}
