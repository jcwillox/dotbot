package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/yamltools"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Config         PluginList
	Profiles       ProfilesBase
	DefaultProfile DefaultProfileBase `yaml:"default_profile"`
	StripPath      StripPathBase      `yaml:"strip_path"`
	Vars           map[string]interface{}
}

func (c *Config) UnmarshalYAML(n *yaml.Node) error {
	err := yamltools.LoadIncludeTag(n)
	if err != nil {
		return err
	}
	err = yamltools.LoadIncludeDirNamedTag(n)
	if err != nil {
		return err
	}
	n = yamltools.ListToMapVal(n, "config")
	type ConfigT Config
	return n.Decode((*ConfigT)(c))
}

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
		"package":  &PackageBase{},
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

func ReadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	return FromBytes(data)
}

func FromBytes(data []byte) (Config, error) {
	config := Config{}
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, err
}

func (c Config) RunAll() {
	template.Vars(c.Vars)
	c.StripPath.Run()
	// groups set via the cli take precedence
	if store.Groups == nil {
		profile := c.DefaultProfile.GetDefaultProfile()
		if profile != "" {
			store.Groups = c.Profiles.GetGroups(profile)
			LogProfile(profile)
		}
	}
	c.Config.RunAll()
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
