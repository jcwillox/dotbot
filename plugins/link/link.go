package link

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
	"path/filepath"
)

type Base []Config

func (b *Base) UnmarshalYAML(n *yaml.Node) error {
	type BaseT Base
	return yamltools.MapSlice(yamltools.EnsureFlatList(n)).Decode((*BaseT)(b))
}

type Config struct {
	Path   string
	Source string
}

func (c *Config) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.KeyValToNamedMap(n, "path", "source")
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
			fmt.Println("error:", err)
		}
	}
	return nil
}

var logger = log.GetLogger(emerald.Cyan, "LINK", emerald.LightBlack)

func (c Config) Run() error {
	sourceStat, err := os.Lstat(c.Source)
	if os.IsNotExist(err) {
		return errors.New("source does not exist")
	}
	path := utils.ExpandUser(c.Path)
	// check if link exists
	pathStat, err := os.Lstat(path)
	if !os.IsNotExist(err) {
		// check is link
		if pathStat.Mode()&os.ModeSymlink == 0 {
			// physical file exists where link wants to be placed
			return nil
		}
		dest, err := os.Readlink(path)
		if err != nil {
			return err
		}
		destStat, _ := os.Lstat(dest)
		// check link is already to correct dest
		if os.SameFile(destStat, sourceStat) {
			// link is correct
			logger.LogPath(
				"exists",
				emerald.HighlightPathStat(c.Path, pathStat),
				emerald.HighlightPathStat(dest, destStat),
			)
		}
	} else {
		// destination doesn't exist so create link
		absSource, _ := filepath.Abs(c.Source)
		if !store.DryRun {
			err := os.Symlink(absSource, c.Path)
			if err != nil {
				return err
			}
		}
		logger.LogPath(
			"created",
			emerald.HighlightPath(c.Path, os.ModeSymlink),
			emerald.HighlightPathStat(absSource, sourceStat),
		)
	}
	return nil
}
