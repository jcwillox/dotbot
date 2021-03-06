package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Config         PluginList
	Profiles       ProfilesBase
	DefaultProfile DefaultProfileBase `yaml:"default_profile"`
	UpdateRepo     *bool              `yaml:"update_repo"`
	UpdateDotbot   *bool              `yaml:"update_dotbot"`
	ShowTotalTime  *bool              `yaml:"show_total_time"`
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

type PluginList []map[string]Plugin

type Plugin interface {
	Enabled() bool
	RunAll() error
}

func getDirective(key string) Plugin {
	plugin, present := map[string]Plugin{
		"clean":    &CleanBase{},
		"create":   &CreateBase{},
		"download": &DownloadBase{},
		"extract":  &ExtractBase{},
		"git":      &GitBase{},
		"group":    &GroupBase{},
		"if":       &IfBase{},
		"install":  &InstallBase{},
		"link":     &LinkBase{},
		"package":  &PackageBase{},
		"sharkdp":  &SharkdpBase{},
		"shell":    &ShellBase{},
		"system":   &SystemBase{},
		"vars":     &VarsBase{},
	}[key]
	if !present {
		return nil
	}
	return plugin
}

func (c *PluginList) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	*c = make(PluginList, len(n.Content))
	for i, node := range n.Content {
		// range over keys
		keys := yamltools.MapKeys(node)
		// lookup concrete type
		plugin := getDirective(keys[0])
		if plugin == nil {
			log.Warnf("skipping unknown directive '%s'\n", keys[0])
			continue
		}
		// decode into type
		err := node.Content[1].Decode(plugin)
		if err != nil {
			return err
		}
		// set index
		(*c)[i] = map[string]Plugin{keys[0]: plugin}
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

// RunAll runs all configs returns true if the config should be reloaded
func (c Config) RunAll(useBasic ...bool) bool {
	store.TmplVars(c.Vars)
	c.StripPath.Run()

	if useBasic == nil {
		if (c.UpdateDotbot == nil || *c.UpdateDotbot == true) && os.Getenv("DOTBOT_NO_UPDATE") != "1" {
			UpdaterUpdate()
		}
		if (c.UpdateRepo == nil || *c.UpdateRepo == true) && os.Getenv("DOTBOT_NO_UPDATE_REPO") != "1" {
			didUpdate, err := UpdaterUpdateRepo()
			if err != nil {
				log.Fatalln("failed to update dotfiles repo:", err)
			}
			if didUpdate {
				return true
			}
		}
	}

	// groups set via the cli take precedence
	if store.Groups == nil {
		profile := c.DefaultProfile.GetDefaultProfile()
		if profile != "" {
			store.Groups = c.Profiles.GetGroups(profile)
			LogProfile(profile)
		}
	}

	if useBasic == nil && (c.ShowTotalTime == nil || *c.ShowTotalTime == true) {
		start := time.Now()
		c.Config.RunAll()
		fmt.Print(
			emerald.ColorCode("cyan+d"), "[", utils.FormatDuration(time.Since(start)), "]", emerald.Reset, "\n",
		)
	} else {
		c.Config.RunAll()
	}

	return false
}

func (c PluginList) RunAll() {
	errorCount := 0
	for _, item := range c {
		for _, plugin := range item {
			if plugin.Enabled() {
				err := plugin.RunAll()
				if err != nil {
					errorCount++
				}
			}
		}
	}
	if errorCount > 0 {
		fmt.Printf("ERROR: %d tasks failed out of %d\n", errorCount, len(c))
	}
	store.RemoveTempFiles()
}
