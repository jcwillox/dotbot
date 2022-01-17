//go:build ignore
// +build ignore

package plugin

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"gopkg.in/yaml.v3"
)

var pluginLogger = log.NewBasicLogger("PLUGIN")

type PluginBase []PluginConfig
type PluginConfig struct {
}

func (b *PluginBase) UnmarshalYAML(n *yaml.Node) error {
	type PluginBaseT PluginBase
	return n.Decode((*PluginBaseT)(b))
}

func (b PluginBase) Enabled() bool {
	return true
}

func (b PluginBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

func (c PluginConfig) Run() error {
	return nil
}
