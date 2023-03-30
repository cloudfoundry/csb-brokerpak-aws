package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-dynamodbns/csbdynamodbns"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbdynamodbns.Provider,
	})
}
