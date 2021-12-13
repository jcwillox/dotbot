package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type CleanBase []*CleanConfig
type CleanConfig struct {
	Path      string `yaml:",omitempty"`
	Force     bool
	Recursive bool
}

func (b *CleanBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapToSliceMap(n)
	n = yamltools.EnsureList(n)
	type CleanBaseT CleanBase
	return n.Decode((*CleanBaseT)(b))
}

func (c *CleanConfig) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureMapMap(n)
	n = yamltools.MapKeyIntoValueMap(n, "path")
	type CleanConfigT CleanConfig
	return n.Decode((*CleanConfigT)(c))
}

func (c *CleanConfig) MarshalYAML() (interface{}, error) {
	path := c.Path
	c.Path = ""
	type CleanConfigT CleanConfig
	return map[string]*CleanConfigT{path: (*CleanConfigT)(c)}, nil
}

func (b CleanBase) Enabled() bool {
	return true
}

func (b CleanBase) RunAll() error {
	paths := make([]string, len(b))
	for i, config := range b {
		paths[i] = emerald.HighlightPath(config.Path, os.ModeDir)
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	cleanLogger.Log().Tag("cleaned").Println(strings.Join(paths, emerald.LightBlack+", "+emerald.Reset))
	return nil
}

var cleanLogger = log.GetLogger(emerald.ColorCode("red+b"), "CLEAN", emerald.Yellow)

func (c CleanConfig) Run() error {
	path := utils.ExpandUser(c.Path)
	err := filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if c.Path == path {
			return nil
		}
		if !c.Recursive && entry.IsDir() {
			return filepath.SkipDir
		}
		// ignore non-symlinks
		if entry.Type()&os.ModeSymlink == 0 {
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
		pathStat, _ := entry.Info()
		// check dead link
		if stat, err := os.Stat(dest); err != nil {
			cleanLogger.LogPath("removing", emerald.HighlightPathStat(utils.ShrinkUser(path), pathStat), emerald.HighlightPathStat(dest, stat))
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
