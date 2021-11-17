package clean

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Base []Config

type Config struct {
	Path      string
	Force     bool
	Recursive bool
}

func (b *Base) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapSlice(n)
	n = yamltools.EnsureList(n)
	type BaseT Base
	return n.Decode((*BaseT)(b))
}

func (c *Config) UnmarshalYAML(n *yaml.Node) error {
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

var logger = log.GetLogger(emerald.Red, "CLEAN", emerald.LightBlack)

func (c Config) Run() error {
	path := utils.ExpandUser(c.Path)
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if c.Path == path {
			return nil
		}
		if !c.Recursive && info.IsDir() {
			return filepath.SkipDir
		}
		// ignore non-symlinks
		if info.Mode()&os.ModeSymlink == 0 {
			return nil
		}
		dest, err := os.Readlink(path)
		if err != nil {
			return err
		}
		// check link is to dotfiles directory
		rel, err := filepath.Rel(store.BaseDirectory, dest)
		if !c.Force && (err != nil || strings.HasPrefix(rel, "..")) {
			return nil
		}
		// check dead link
		if stat, err := os.Stat(dest); err != nil {
			logger.LogPath("removing", emerald.HighlightPathStat(utils.ShrinkUser(path), info), emerald.HighlightPathStat(dest, stat))
			if !store.DryRun {
				return os.Remove(path)
			}
		}
		return nil
	})
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
