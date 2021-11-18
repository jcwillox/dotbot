package git

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/go-git/go-git/v5"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"io"
	"path/filepath"
)

type Base []Config

type Config struct {
	Path string
	Url  string
	Name string
	// one of clone, pull, clone_pull
	Method  string `default:"clone_pull"`
	Shallow bool   `default:"true"`
}

func (b *Base) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapSlice(n)
	type BaseT Base
	return n.Decode((*BaseT)(b))
}

func (c *Config) UnmarshalYAML(n *yaml.Node) error {
	defaults.MustSet(c)
	n = yamltools.EnsureMapMap(n)
	n = yamltools.KeyMapToNamedMap(n, "path")
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

var logger = log.GetLogger(emerald.ColorCode("yellow+b"), "GIT", emerald.Yellow)

func (c Config) Run() error {
	path := utils.ExpandUser(c.Path)
	_, err := git.PlainOpen(path)
	isNotExists := errors.Is(err, git.ErrRepositoryNotExists)
	if !isNotExists && err != nil {
		return err
	}

	// check if we can write to the parent directory
	sudo := !utils.IsWritable(filepath.Dir(path))

	logTail := func() {
		logger.Sudo(sudo).Print(emerald.Blue, c.String(), "\n")
	}

	switch c.Method {
	case "clone_pull":
		if isNotExists {
			logger.Log().Tag("cloning")
			logTail()
			return c.clonePath(path, sudo)
		}
		logger.Log().Tag("pulling")
		logTail()
		return c.pullPath(path, sudo)
	case "clone":
		if isNotExists {
			logger.Log().Tag("cloning")
			logTail()
			return c.clonePath(path, sudo)
		} else {
			logger.LogTagC(emerald.LightBlack, "cloned", emerald.Blue, c)
		}
	case "pull":
		if isNotExists {
			return err
		}
		logger.Log().Tag("pulling")
		logTail()
		return c.pullPath(path, sudo)
	}
	return nil
}

func (c Config) String() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Url
}

func (c Config) clonePath(path string, sudo bool) error {
	if store.DryRun {
		return nil
	}
	flags := ""
	if c.Shallow {
		flags = "--depth=1"
	}
	cmd, err := utils.Command{
		Command: fmt.Sprintln("git clone", flags, c.Url, path),
		Shell:   false,
		Stdout:  true,
		Stderr:  true,
		Sudo:    sudo,
	}.Cmd()
	if err != nil {
		return err
	}
	return cmd.Run()
}

func (c Config) pullPath(path string, sudo bool) error {
	if store.DryRun {
		return nil
	}
	cmd, err := utils.Command{
		Command: fmt.Sprintf("git -c color.ui=always -C %s pull --progress", path),
		Shell:   false,
		Stderr:  true,
		Sudo:    sudo,
	}.Cmd()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(stdout)
	err = cmd.Start()
	if err != nil {
		return err
	}
	for {
		out, err := reader.ReadString('\n')
		if out == "Already up to date.\n" {
			if emerald.ColorEnabled {
				emerald.Print("\x1b[F\x1b[K")
			}
			logger.LogTagC(emerald.LightBlack, "up-to-date", emerald.Blue, c)
		} else {
			emerald.Print(out)
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
