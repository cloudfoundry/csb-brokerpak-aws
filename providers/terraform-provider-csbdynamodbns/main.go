package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.gwd.broadcom.net/TNZ/csb-brokerpak-aws/terraform-provider-dynamodbns/csbdynamodbns"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbdynamodbns.Provider,
	})
}
