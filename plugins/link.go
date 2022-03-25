package plugins

import (
	"errors"
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
	"os"
	"path/filepath"
	"strconv"
)

var linkLogger = log.NewBasicLogger("LINK")

type LinkBase []*LinkConfig
type LinkConfig struct {
	Path      string `yaml:",omitempty"`
	Source    string
	Mkdirs    bool `default:"true"`
	Force     bool
	SafeForce bool `yaml:"safe_force"`
}

func (b *LinkBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureFlatList(n)
	n = yamltools.MapToSliceMap(n)
	type LinkBaseT LinkBase
	return n.Decode((*LinkBaseT)(b))
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
		if sudo.IsPermission(err) && sudo.WouldSudo() {
			absSource, _ := filepath.Abs(config.Source)
			if !sudo.HasUsedSudo {
				linkLogger.TagSudo("linking", true).Path(
					emerald.HighlightPath(config.Path, os.ModeSymlink),
					emerald.HighlightPathStat(absSource),
				)
			}
			err = sudo.Config("link", &config)
		}
		if err != nil {
			fmt.Println("error:", err)
		}
	}
	return nil
}

func (c LinkConfig) Run() error {
	err := template.RenderField(&c.Path, &c.Source)
	if err != nil {
		return err
	}

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
				linkLogger.TagDone("linked").Path(
					emerald.HighlightPathStat(c.Path, pathStat),
					emerald.HighlightPathStat(dest, destStat),
				)
				return nil
			}
		}
		if c.Force || c.SafeForce {
			if !utils.IsWritable(path) {
				return os.ErrPermission
			}
			if c.Force {
				if !store.DryRun {
					err := os.Remove(path)
					if err != nil {
						return err
					}
				}
				linkLogger.TagC(emerald.Red, "deleted").Println(emerald.HighlightPathStat(c.Path, pathStat))
			} else {
				for i := 1; i < 11; i++ {
					dest := path + "." + strconv.Itoa(i)
					if _, err := os.Lstat(dest); os.IsNotExist(err) {
						if !store.DryRun {
							err := os.Rename(path, dest)
							if err != nil {
								return err
							}
						}
						linkLogger.TagC(emerald.Red, "renamed").Path(
							emerald.HighlightPathStat(c.Path, pathStat),
							emerald.HighlightPathStat(c.Path+"."+strconv.Itoa(i), pathStat),
						)
						break
					}
					if i == 10 {
						return errors.New("unable to rename file: too many failed renames")
					}
				}
			}
		} else {
			return errors.New("failed to create link as target already exists")
		}
	}

	// at this point the target does not exist
	if c.Mkdirs {
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
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

	linkLogger.TagSudo("linked").Path(
		emerald.HighlightPath(c.Path, os.ModeSymlink),
		emerald.HighlightPathStat(absSource, sourceStat),
	)
	return nil
}
