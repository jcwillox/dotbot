package plugins

import (
	"bytes"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/utils/sudo"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"io"
)

var shellLogger = log.NewBasicLogger("SHELL")

type ShellBase []*ShellConfig
type ShellConfig struct {
	Desc    string
	Command utils.Command `yaml:",inline"`
	Capture bool
}

func (b *ShellBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type ShellBaseT ShellBase
	return n.Decode((*ShellBaseT)(b))
}

func (c *ShellConfig) UnmarshalYAML(n *yaml.Node) error {
	defaults.MustSet(c)
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

func (c ShellConfig) Run() error {
	err := template.RenderField(&c.Command.Command)
	if err != nil {
		return err
	}
	if c.Desc != "false" {
		willSudo := (c.Command.Sudo || c.Command.TrySudo) && sudo.WouldSudo()
		if c.Desc == "" {
			shellLogger.TagSudo("running", willSudo).Print(emerald.LightBlue, c.Command.ShortString(), "\n")
		} else {
			shellLogger.TagSudo("running", willSudo).Print(
				emerald.LightBlue, c.Desc, " ", emerald.LightBlack, "'", c.Command.ShortString(), "'\n",
			)
		}
	}
	cmd, err := c.Command.Cmd()
	if err != nil || store.DryRun {
		return err
	}

	// support capturing stdout/stderr
	if c.Capture {
		var stdout, stderr bytes.Buffer
		if cmd.Stdout != nil {
			cmd.Stdout = io.MultiWriter(&stdout, cmd.Stdout)
		} else {
			cmd.Stdout = &stdout
		}
		if cmd.Stderr != nil {
			cmd.Stderr = io.MultiWriter(&stderr, cmd.Stdout)
		} else {
			cmd.Stderr = &stderr
		}
		// add output to template variables after execution
		defer func() {
			store.TmplVars(map[string]interface{}{
				"Stdout": stdout.String(),
				"Stderr": stderr.String(),
			})
		}()
	}

	return cmd.Run()
}
