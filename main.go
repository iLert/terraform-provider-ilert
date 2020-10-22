package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/iLert/terraform-provider-ilert/ilert"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ilert.Provider,
	})
}
