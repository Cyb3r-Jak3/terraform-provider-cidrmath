package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/Cyb3r-Jak3/terraform-provider-cidrmath/internal/provider"
)

var (
	version    = "dev"
	commit     = "none"
	commitDate = "unknown"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version, commitDate, commit), providerserver.ServeOpts{
		Address: "registry.terraform.io/Cyb3r-Jak3/cidrmath",
	})
	if err != nil {
		log.Fatal(err)
	}
}
