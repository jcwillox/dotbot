package plugins

import (
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"strings"
)

type ProfilesBase []ProfileConfig

type ProfileConfig struct {
	Name   string
	Groups FlatList
}

func (b *ProfilesBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapToSliceMap(n)
	type ProfilesBaseT ProfilesBase
	return n.Decode((*ProfilesBaseT)(b))
}

func (c *ProfileConfig) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapSplitKeyVal(n, "name", "groups")
	type ProfileConfigT ProfileConfig
	return n.Decode((*ProfileConfigT)(c))
}

type DefaultProfileBase []DefaultProfileConfig
type DefaultProfileConfig struct {
	Profile  string
	Template string
}

func (b *DefaultProfileBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.ScalarToList(n)
	n = yamltools.MapToSliceMap(n)
	type DefaultProfileBaseT DefaultProfileBase
	return n.Decode((*DefaultProfileBaseT)(b))
}

func (c *DefaultProfileConfig) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.ScalarToMap(n)
	n = yamltools.MapSplitKeyVal(n, "profile", "template")
	type DefaultProfileConfigT DefaultProfileConfig
	return n.Decode((*DefaultProfileConfigT)(c))
}

type FlatList []string

func (l *FlatList) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	n = yamltools.EnsureFlatList(n)
	type FlatListT FlatList
	return n.Decode((*FlatListT)(l))
}

func (b ProfilesBase) GetGroups(profile string) []string {
	for _, config := range b {
		if config.Name == profile {
			return config.Groups
		}
	}
	return nil
}

func (b DefaultProfileBase) GetDefaultProfile() string {
	for _, config := range b {
		if config.Template == "" {
			return config.Profile
		}
		result, err := template.Parse(config.Template).Render()
		if err != nil {
			log.Fatalln("Failed to render profile template", err)
		}
		if strings.EqualFold(result, "true") {
			return config.Profile
		}
	}
	return ""
}

var profileLogger = log.GetLogger(emerald.ColorCode("magenta+b"), "PROFILE", emerald.Red)

func LogProfile(name string) {
	profileLogger.Log().Print(
		emerald.Red, name, emerald.Reset, "; ", emerald.Cyan,
		strings.Join(store.Groups, emerald.White+", "+emerald.Cyan), "\n",
	)
}
