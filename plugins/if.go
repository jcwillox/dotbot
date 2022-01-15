package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/yamltools"
	"gopkg.in/yaml.v3"
)

type IfBase []IfConfig
type IfConfig struct {
	Condition FlatList
	Then      PluginList
	Else      PluginList
}

func (b *IfBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type IfBaseT IfBase
	return n.Decode((*IfBaseT)(b))
}
func (c *IfConfig) UnmarshalYAML(n *yaml.Node) error {
	type IfConfigT IfConfig
	return n.Decode((*IfConfigT)(c))
}

func (b IfBase) Enabled() bool {
	return true
}

func (b IfBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

func (c IfConfig) Run() error {
	for _, condition := range c.Condition {
		result, err := template.Parse(condition).RenderTrue()
		if err != nil {
			return err
		}
		if !result {
			c.Else.RunAll()
			return nil
		}
	}
	c.Then.RunAll()
	return nil
}
