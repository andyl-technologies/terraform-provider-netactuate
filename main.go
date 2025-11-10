package main

import (
	"flag"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/netactuate/terraform-provider-netactuate/netactuate"
)

const (
	ProviderAddr = "registry.terraform.io/netactuate/netactuate"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: netactuate.Provider,
		ProviderAddr: ProviderAddr,

		Debug: debugMode,
		Logger: hclog.FromStandardLogger(
			log.Default(),
			&hclog.LoggerOptions{
				Name: "terraform-provider-netactuate",
			},
		),
	}

	plugin.Serve(opts)
}
