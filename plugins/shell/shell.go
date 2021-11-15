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
	if c.Desc == "" {
		logger.Log(c.Command.ShortString())
	} else {
		logger.Log(c.Desc, emerald.LightBlack, " ["+c.Command.ShortString()+"]")
	}

	cmd, err := c.Command.Cmd()
	if err != nil {
		return err
	}
	return cmd.Run()
}
