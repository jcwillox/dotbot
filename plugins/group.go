package plugins

import (
	"bytes"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
)

type GroupBase []GroupConfig
type GroupConfig struct {
	Name   string
	Config PluginList
}

func (b *GroupBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapToSliceMap(n)
	type GroupBaseT GroupBase
	return n.Decode((*GroupBaseT)(b))
}

func (c *GroupConfig) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapSplitKeyVal(n, "name", "config")
	type GroupConfigT GroupConfig
	return n.Decode((*GroupConfigT)(c))
}

func (b GroupBase) Enabled() bool {
	return true
}

func logGroup(group string) {
	freeWidth := 60 - len(group) - 2
	bar := string(bytes.Repeat([]byte("─"), freeWidth/2))
	emerald.Print(
		emerald.LightBlack, bar,
		emerald.Bold, emerald.Green, " ", group, " ", emerald.Reset,
		emerald.LightBlack, bar, emerald.Reset, "\n",
	)
}

func (b GroupBase) RunAll() error {
	for _, c := range b {
		if store.Groups != nil {
			for _, group := range store.Groups {
				if c.Name == group {
					logGroup(group)
					c.Config.RunAll()
				}
			}
		} else {
			logGroup(c.Name)
			c.Config.RunAll()
		}

	}
	return nil
}
