package plugins

import (
	"errors"
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"os"
)

type CreateBase []CreateConfig

func (b *CreateBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapSlice(n)
	n = yamltools.EnsureList(n)
	type CreateBaseT CreateBase
	return n.Decode((*CreateBaseT)(b))
}

type CreateConfig struct {
	Path string `yaml:",omitempty"`
	Mode os.FileMode
}

func (c *CreateConfig) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureMap(n)
	n = yamltools.KeyValToNamedMap(n, "path", "mode")
	n = yamltools.KeyMapToNamedMap(n, "path")
	type CreateConfigT CreateConfig
	err := n.Decode((*CreateConfigT)(c))
	if err != nil {
		return err
	}
	if c.Mode == 0 {
		c.Mode = 0755
	}
	c.Mode |= os.ModeDir
	return nil
}

func (c *CreateConfig) MarshalYAML() (interface{}, error) {
	path := c.Path
	c.Path = ""
	type CreateConfigT CreateConfig
	return map[string]*CreateConfigT{path: (*CreateConfigT)(c)}, nil
}

func (b CreateBase) Enabled() bool {
	return true
}

var nonExistentPath = emerald.ColorFunc("red+u")

func (b CreateBase) RunAll() error {
	hasError := false
	for _, config := range b {
		err := config.Run()
		if utils.IsPermError(err) && utils.WouldSudo() {
			// let user know why we want to sudo
			createLogger.Log().TagC(emerald.Yellow, "creating").Sudo(true).Print(
				emerald.HighlightFileMode(config.Mode), " ", emerald.HighlightPath(config.Path, os.ModeDir), "\n",
			)
			return utils.SudoConfig("create", &config)
		}
		if err != nil {
			log.Error("Failed to create directory:", nonExistentPath(config.Path))
			fmt.Println(err)
			hasError = true
		}
	}
	if hasError {
		return errors.New("failed to create some directories")
	}
	return nil
}

var createLogger = log.GetLogger(emerald.ColorCode("green+b"), "CREATE", emerald.Yellow)

func (c CreateConfig) Run() error {
	path := utils.ExpandUser(c.Path)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if !store.DryRun {
			err := os.MkdirAll(path, c.Mode)
			if err != nil {
				return err
			}
		}
		createLogger.Log().TagC(emerald.Yellow, "created").Sudo().Print(
			emerald.HighlightFileMode(c.Mode), " ", emerald.HighlightPath(c.Path, os.ModeDir),
		).Println()
	} else if err != nil {
		return err
	} else {
		createLogger.LogTagC(emerald.LightBlack, "exists", emerald.HighlightPath(c.Path, os.ModeDir))
	}
	return nil
}
