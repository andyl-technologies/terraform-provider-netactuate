package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/netactuate/terraform-provider-netactuate/netactuate"
)

const (
	ProviderAddr = "registry.terraform.io/netactuate/netactuate"
	version      = "dev"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	ctx := context.Background()

	// Wrap the SDK v2 provider with tf5to6server to upgrade it to protocol 6
	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		ctx,
		netactuate.NewSDKProvider(version).GRPCProvider,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create the new Plugin Framework provider
	frameworkProvider := providerserver.NewProtocol6(
		netactuate.NewFrameworkProvider(version),
	)

	// Mux the providers together
	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
		frameworkProvider,
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		ProviderAddr,
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
