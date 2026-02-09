package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/iLert/terraform-provider-ilert/v2/ilert"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ilert.Provider,
	})
}
