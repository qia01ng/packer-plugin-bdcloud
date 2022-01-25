package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	ecsbuilder "github.com/qia01ng/packer-plugin-bdcloud/builder/ecs"
	version "github.com/qia01ng/packer-plugin-bdcloud/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("ecs", new(ecsbuilder.Builder))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
