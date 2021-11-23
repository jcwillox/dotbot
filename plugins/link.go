package plugins

import (
	"errors"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type LinkBase []LinkConfig

func (b *LinkBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureFlatList(n)
	n = yamltools.MapToSliceMap(n)
	type LinkBaseT LinkBase
	return n.Decode((*LinkBaseT)(b))
}

type LinkConfig struct {
	Path   string `yaml:",omitempty"`
	Source string
	Mkdirs bool `default:"true"`
	Force  bool
}

func (c *LinkConfig) UnmarshalYAML(n *yaml.Node) error {
	defaults.MustSet(c)
	if yamltools.IsScalarMap(n) {
		n = yamltools.MapSplitKeyVal(n, "path", "source")
	} else {
		n = yamltools.MapKeyIntoValueMap(n, "path")
	}
	type LinkConfigT LinkConfig
	return n.Decode((*LinkConfigT)(c))
}

func (c *LinkConfig) MarshalYAML() (interface{}, error) {
	path := c.Path
	c.Path = ""
	type LinkConfigT LinkConfig
	return map[string]*LinkConfigT{path: (*LinkConfigT)(c)}, nil
}

func (b LinkBase) Enabled() bool {
	return true
}

func (b LinkBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if utils.IsPermError(err) && utils.WouldSudo() {
			absSource, _ := filepath.Abs(config.Source)
			if !utils.HasUsedSudo {
				linkLogger.Log().Tag("linking").Sudo(true).Path(
					emerald.HighlightPath(config.Path, os.ModeSymlink),
					emerald.HighlightPathStat(absSource),
				)
			}
			return utils.SudoConfig("link", &config)
		} else if err != nil {
			fmt.Println("error:", err)
		}
	}
	return nil
}

var linkLogger = log.GetLogger(emerald.ColorCode("cyan+b"), "LINK", emerald.Yellow)

func (c LinkConfig) Run() error {
	sourceStat, err := os.Lstat(c.Source)
	if os.IsNotExist(err) {
		return errors.New("source does not exist")
	}
	path := utils.ExpandUser(c.Path)
	// check if link exists
	pathStat, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return err // general stat error
	}
	if err == nil {
		// target exists
		// check if physical file exists where link wants to be placed
		if pathStat.Mode()&os.ModeSymlink != 0 {
			// check if link is already correct
			dest, err := os.Readlink(path)
			if err != nil {
				return err
			}
			destStat, err := os.Lstat(dest)
			if err != nil && !os.IsNotExist(err) {
				return err // general stat error
			}
			// check link is already to correct dest
			if os.SameFile(destStat, sourceStat) {
				// link is correct
				linkLogger.LogPathC(
					emerald.LightBlack,
					"valid",
					emerald.HighlightPathStat(c.Path, pathStat),
					emerald.HighlightPathStat(dest, destStat),
				)
				return nil
			}
		}
		if c.Force {
			if !utils.IsWritable(path) {
				return os.ErrPermission
			}
			err := os.Remove(path)
			if err != nil {
				return err
			}
			linkLogger.Log().TagC(emerald.Red, "deleted").Println(emerald.HighlightPathStat(c.Path, pathStat))
		} else {
			return errors.New("failed to create link as target already exists")
		}
	}

	// at this point the target does not exist
	if c.Mkdirs {
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}
	}

	// check we can create links in the directory
	if !utils.IsWritable(filepath.Dir(path)) {
		return os.ErrPermission
	}

	absSource, _ := filepath.Abs(c.Source)
	if !store.DryRun {
		err := os.Symlink(absSource, path)
		if err != nil {
			return err
		}
	}

	linkLogger.Log().Tag("linked").Sudo().Path(
		emerald.HighlightPath(c.Path, os.ModeSymlink),
		emerald.HighlightPathStat(absSource, sourceStat),
	)
	return nil
}
