package plugins

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

var gitLogger = log.NewBasicLogger("GIT")

type GitBase []*GitConfig

type GitConfig struct {
	Path string `yaml:",omitempty"`
	Url  string
	Name string `yaml:",omitempty"`
	// one of clone, pull, clone_pull
	Method  string `default:"clone_pull"`
	Shallow bool   `default:"true"`
}

func (b *GitBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapToSliceMap(n)
	type GitBaseT GitBase
	return n.Decode((*GitBaseT)(b))
}

func (c *GitConfig) UnmarshalYAML(n *yaml.Node) error {
	defaults.MustSet(c)
	if yamltools.IsScalarMap(n) {
		n = yamltools.MapSplitKeyVal(n, "path", "url")
	} else {
		n = yamltools.MapKeyIntoValueMap(n, "path")
	}
	type GitConfigT GitConfig
	return n.Decode((*GitConfigT)(c))
}

func (c *GitConfig) MarshalYAML() (interface{}, error) {
	path := c.Path
	c.Path = ""
	type GitConfigT GitConfig
	return map[string]*GitConfigT{path: (*GitConfigT)(c)}, nil
}

func (b GitBase) Enabled() bool {
	return true
}

func (b GitBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

func (c GitConfig) Run() error {
	path := utils.ExpandUser(c.Path)
	_, err := git.PlainOpen(path)
	isNotExists := errors.Is(err, git.ErrRepositoryNotExists)
	if !isNotExists && err != nil {
		return err
	}

	// check if we can write to the parent directory
	sudo := !utils.IsWritable(filepath.Dir(path))

	logAction := func(tag string) {
		gitLogger.TagSudo(tag).Print(emerald.LightBlue, c, "\n")
	}

	switch c.Method {
	case "clone_pull":
		if isNotExists {
			logAction("cloning")
			return c.clonePath(path, sudo)
		}
		logAction("pulling")
		return c.pullPath(path, sudo)
	case "clone":
		if isNotExists {
			logAction("cloning")
			return c.clonePath(path, sudo)
		} else {
			gitLogger.TagDone("cloned").Print(emerald.LightBlue, c, "\n")
		}
	case "pull":
		if isNotExists {
			return err
		}
		logAction("pulling")
		return c.pullPath(path, sudo)
	}
	return nil
}

func (c GitConfig) String() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Url
}

func (c GitConfig) clonePath(path string, sudo bool) error {
	if store.DryRun {
		return nil
	}
	flags := ""
	if c.Shallow {
		flags = "--depth=1"
	}
	cmd, err := utils.Command{
		Command: fmt.Sprintf("git clone %s '%s' '%s'", flags, c.Url, path),
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

var DidGitUpdate bool

func (c GitConfig) pullPath(path string, sudo bool) error {
	DidGitUpdate = true
	if store.DryRun {
		return nil
	}
	cmd, err := utils.Command{
		Command: fmt.Sprintf("git -c color.ui=always -C '%s' pull --progress", path),
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
			DidGitUpdate = false
			if emerald.ColorEnabled {
				emerald.Print("\x1b[F\x1b[K")
			}
			gitLogger.TagDone("up-to-date").Print(emerald.LightBlue, c, "\n")
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
