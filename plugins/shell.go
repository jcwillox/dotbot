package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
)

type ShellBase []ShellConfig

func (b *ShellBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type ShellBaseT ShellBase
	return n.Decode((*ShellBaseT)(b))
}

type ShellConfig struct {
	Desc    string
	Command utils.Command `yaml:",inline"`
}

func (c *ShellConfig) UnmarshalYAML(n *yaml.Node) error {
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
	type ShellConfigT ShellConfig
	return n.Decode((*ShellConfigT)(c))
}

func (b ShellBase) Enabled() bool {
	return true
}

func (b ShellBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

var shellLogger = log.GetLogger(emerald.Magenta, "SHELL", emerald.LightBlack)

func (c ShellConfig) Run() error {
	shellLogger.Log()
	if c.Desc == "" {
		shellLogger.Print(emerald.Yellow, c.Command.ShortString(), " ")
		if c.Command.Sudo || (c.Command.TrySudo && utils.CanSudo()) {
			shellLogger.TagC(emerald.Blue, "sudo")
		}
		shellLogger.Println()
	} else {
		shellLogger.Print(emerald.Yellow, c.Desc, " ")
		if c.Command.Sudo || (c.Command.TrySudo && utils.CanSudo()) {
			shellLogger.TagC(emerald.Blue, "sudo")
		}
		shellLogger.Print(emerald.LightBlack, "[", c.Command.ShortString(), "]\n")

	}

	cmd, err := c.Command.Cmd()
	if err != nil {
		return err
	}
	return cmd.Run()
}
