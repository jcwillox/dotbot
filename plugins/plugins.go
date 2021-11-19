package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/yamltools"
	"gopkg.in/yaml.v3"
	"os"
)

type PluginList []Plugin

type Plugin interface {
	Enabled() bool
	RunAll() error
}

func getDirective(key string) Plugin {
	return map[string]Plugin{
		"clean":    &CleanBase{},
		"create":   &CreateBase{},
		"git":      &GitBase{},
		"group":    &GroupBase{},
		"download": &DownloadBase{},
		"link":     &LinkBase{},
		"shell":    &ShellBase{},
	}[key]
}

func (c *PluginList) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	*c = make(PluginList, len(n.Content))
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

func ReadConfig(path string) (PluginList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return FromBytes(data)
}

func FromBytes(data []byte) (PluginList, error) {
	config := make(PluginList, 0, 5)
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, err
}

func (c PluginList) RunAll() {
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
