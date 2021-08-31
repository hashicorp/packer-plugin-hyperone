package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	"github.com/hashicorp/packer-plugin-hyperone/builder/hyperone"
	"github.com/hashicorp/packer-plugin-hyperone/version"
)

func main() {
	pps := plugin.NewSet()
	pps.SetVersion(version.PluginVersion)
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(hyperone.Builder))
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
