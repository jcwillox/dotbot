package shell

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
)

type Base []Config

func (b *Base) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type BaseT Base
	return n.Decode((*BaseT)(b))
}

type Config struct {
	Desc    string
	Command utils.Command `yaml:",inline"`
}

func (c *Config) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		n = &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
			Content: []*yaml.Node{{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "command",
			}, n},
		}
	}
	type ConfigT Config
	return n.Decode((*ConfigT)(c))
}

func (b Base) Enabled() bool {
	return true
}

func (b Base) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

var logger = log.GetLogger(emerald.Magenta, "SHELL", emerald.LightBlack)

func (c Config) Run() error {
	logger.Log()
	if c.Desc == "" {
		logger.Print(emerald.Yellow, c.Command.ShortString(), " ")
		if c.Command.Sudo || (c.Command.TrySudo && utils.CanSudo()) {
			logger.TagC(emerald.Blue, "sudo")
		}
		logger.Println()
	} else {
		logger.Print(emerald.Yellow, c.Desc, " ")
		if c.Command.Sudo || (c.Command.TrySudo && utils.CanSudo()) {
			logger.TagC(emerald.Blue, "sudo")
		}
		logger.Print(emerald.LightBlack, "[", c.Command.ShortString(), "]\n")

	}

	cmd, err := c.Command.Cmd()
	if err != nil {
		return err
	}
	return cmd.Run()
}
