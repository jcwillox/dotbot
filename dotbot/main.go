package main

import (
	"github.com/jcwillox/dotbot/dotbot/cmd"
	"github.com/jcwillox/dotbot/plugins"
	"os"
)

func main() {
	plugins.UpdaterCleanup()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
