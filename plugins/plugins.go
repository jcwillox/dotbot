package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/plugins/clean"
	"github.com/jcwillox/dotbot/plugins/create"
	"github.com/jcwillox/dotbot/plugins/download"
	"github.com/jcwillox/dotbot/plugins/link"
	"github.com/jcwillox/dotbot/plugins/shell"
	"github.com/jcwillox/dotbot/yamltools"
	"gopkg.in/yaml.v3"
	"os"
)

type Config []Plugin

func getDirective(key string) Plugin {
	return map[string]Plugin{
		"clean":    &clean.Base{},
		"create":   &create.Base{},
		"download": &download.Base{},
		"link":     &link.Base{},
		"shell":    &shell.Base{},
	}[key]
}

func (c *Config) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	*c = make(Config, len(n.Content))
	for i, node := range n.Content {
		// range over keys
		keys := yamltools.MapKeys(node)
		// lookup concrete type
		plugin := getDirective(keys[0])
		// decode into type
		err := node.Content[1].Decode(plugin)
		if err != nil {
			return err
		}
		// set index
		(*c)[i] = plugin
	}
	return nil
}

type Plugin interface {
	Enabled() bool
	RunAll() error
}

func ReadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return FromBytes(data)
}

func FromBytes(data []byte) (Config, error) {
	config := make(Config, 0, 5)
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, err
}

func (c Config) RunAll() {
	errorCount := 0
	for _, plugin := range c {
		err := plugin.RunAll()
		if err != nil {
			errorCount++
		}
	}
	if errorCount > 0 {
		fmt.Printf("ERROR: %d tasks failed out of %d\n", errorCount, len(c))
	}
}
