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
type ShellConfig struct {
	Desc    string
	Command utils.Command `yaml:",inline"`
}

func (b *ShellBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type ShellBaseT ShellBase
	return n.Decode((*ShellBaseT)(b))
}

func (c *ShellConfig) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.ScalarToMapVal(n, "command")
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

var shellLogger = log.GetLogger(emerald.ColorCode("magenta+b"), "SHELL", emerald.LightBlack)

func (c ShellConfig) Run() error {
	logSudo := func() {
		shellLogger.Sudo((c.Command.Sudo || c.Command.TrySudo) && utils.WouldSudo())
	}
	if c.Desc == "" {
		shellLogger.Log().Print(emerald.Blue, c.Command.ShortString(), " ")
		logSudo()
		shellLogger.Println()
	} else {
		shellLogger.Log().Print(emerald.Blue, c.Desc, " ")
		logSudo()
		shellLogger.Print(emerald.LightBlack, "[", c.Command.ShortString(), "]\n")
	}
	cmd, err := c.Command.Cmd()
	if err != nil {
		return err
	}
	return cmd.Run()
}
