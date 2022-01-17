package plugins

import (
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/yamltools"
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
	err := n.Decode((*GroupConfigT)(c))
	store.RegisteredGroups = append(store.RegisteredGroups, c.Name)
	return err
}

func (b GroupBase) Enabled() bool {
	return true
}

func logGroup(group string) {
	log.Rule(group)
}

func (b GroupBase) RunAll() error {
	if store.Groups != nil {
		for _, group := range store.Groups {
			for _, c := range b {
				if c.Name == group {
					logGroup(group)
					c.Config.RunAll()
				}
			}

		}
	} else {
		for _, c := range b {
			logGroup(c.Name)
			c.Config.RunAll()
		}
	}
	return nil
}
