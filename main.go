package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/jtopjian/terraform-provider-lxc/lxc"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: lxc.Provider,
	})
}
