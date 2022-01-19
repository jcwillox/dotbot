package plugins

import (
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/utils/sudo"
	"github.com/jcwillox/dotbot/yamltools"
	"gopkg.in/yaml.v3"
	"runtime"
)

type SystemBase []SystemConfig
type SystemConfig struct {
	OS       string
	Arch     string
	Platform string
	Family   string
	Libc     string
	Distro   string
	IsRoot   bool `yaml:"is_root"`
	CanSudo  bool `yaml:"can_sudo"`
	Then     PluginList
}

func (b *SystemBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type SystemBaseT SystemBase
	return n.Decode((*SystemBaseT)(b))
}

func (b SystemBase) Enabled() bool {
	return true
}

func (b SystemBase) RunAll() error {
	for _, config := range b {
		if config.Run() {
			return nil
		}
	}
	return nil
}

func (c SystemConfig) Run() bool {
	if c.OS != "" && c.OS != runtime.GOOS {
		return false
	}
	if c.Arch != "" && c.Arch != runtime.GOARCH {
		return false
	}
	if c.Platform != "" || c.Family != "" {
		platform, family := utils.GetPlatformInfo()
		if c.Platform != "" && c.Platform != platform {
			return false
		}
		if c.Family != "" && c.Family != family {
			return false
		}
	}
	if c.Libc != "" && c.Libc != utils.GetLibc() {
		return false
	}
	if c.Distro != "" && c.Distro != template.Distro() {
		return false
	}
	if c.IsRoot && !sudo.IsRoot() {
		return false
	}
	if c.CanSudo && !sudo.CanSudo() {
		return false
	}
	c.Then.RunAll()
	return true
}
