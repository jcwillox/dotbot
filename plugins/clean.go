package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/utils/sudo"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

var cleanLogger = log.NewBasicLogger("CLEAN")

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
	cleaned := false
	for i, config := range b {
		paths[i] = emerald.HighlightPath(config.Path, os.ModeDir)
		cleaned_, err := config.Run()
		if sudo.IsPermission(err) && sudo.WouldSudo() {
			if !sudo.HasUsedSudo {
				linkLogger.TagSudo("cleaning", true).Println(paths[i])
			}
			err = sudo.Config("clean", &config)
		}
		if err != nil {
			fmt.Println("error:", err)
		}
		if cleaned_ == true {
			cleaned = true
		}
	}
	if cleaned {
		cleanLogger.Tag("cleaned").Println(strings.Join(paths, emerald.LightBlack+", "+emerald.Reset))
	} else {
		cleanLogger.TagDone("cleaned").Println(strings.Join(paths, emerald.LightBlack+", "+emerald.Reset))
	}
	return nil
}

func (c CleanConfig) Run() (bool, error) {
	cleaned := false
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
		rel, err := filepath.Rel(store.BaseDir(), dest)
		if !c.Force && (err != nil || strings.HasPrefix(rel, "..")) {
			return nil
		}
		pathStat, _ := entry.Info()
		// check dead link
		if stat, err := os.Stat(dest); err != nil {
			cleaned = true
			if !store.DryRun {
				err := os.Remove(path)
				if err != nil {
					return err
				}
			}
			cleanLogger.TagC(emerald.Red, "deleted").Path(
				emerald.HighlightPathStat(utils.ShrinkUser(path), pathStat),
				emerald.HighlightPathStat(dest, stat),
			)
		}
		return nil
	})
	if os.IsNotExist(err) {
		return cleaned, nil
	}
	return cleaned, err
}
