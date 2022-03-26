package plugins

import (
	"archive/tar"
	"errors"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"github.com/klauspost/compress/zip"
	"github.com/mholt/archiver/v3"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var extractLogger = log.NewBasicLogger("EXTRACT")

type ExtractBase []*ExtractConfig
type ExtractConfig struct {
	Archive string `yaml:",omitempty"`
	Items   ExtractItems
}

type ExtractItems []*ExtractItem
type ExtractItem struct {
	Source  string `yaml:",omitempty"`
	Path    string
	Strip   int
	Replace bool
}

func (b *ExtractBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapToSliceMap(n)
	type ExtractBaseT ExtractBase
	return n.Decode((*ExtractBaseT)(b))
}

func (c *ExtractConfig) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapSplitKeyVal(n, "archive", "items")
	type ExtractConfigT ExtractConfig
	return n.Decode((*ExtractConfigT)(c))
}

func (c *ExtractConfig) MarshalYAML() (interface{}, error) {
	archive := c.Archive
	c.Archive = ""
	return map[string]ExtractItems{archive: c.Items}, nil
}

func (c *ExtractItems) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapToSliceMap(n)
	type ExtractItemsT ExtractItems
	return n.Decode((*ExtractItemsT)(c))
}

func (c *ExtractItem) UnmarshalYAML(n *yaml.Node) error {
	if yamltools.IsScalarMap(n) {
		n = yamltools.MapSplitKeyVal(n, "source", "path")
	} else {
		n = yamltools.MapKeyIntoValueMap(n, "source")
	}
	type ExtractItemT ExtractItem
	return n.Decode((*ExtractItemT)(c))
}

func (c *ExtractItem) MarshalYAML() (interface{}, error) {
	source := c.Source
	c.Source = ""
	type ExtractItemT ExtractItem
	return map[string]*ExtractItemT{source: (*ExtractItemT)(c)}, nil
}

func (b ExtractBase) Enabled() bool {
	return true
}

func (b ExtractBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

func (c ExtractConfig) Run() error {
	err := template.RenderField(&c.Archive)
	if err != nil {
		return err
	}
	archive := utils.ExpandUser(c.Archive)

	f, err := archiver.ByExtension(archive)
	if err != nil {
		return err
	}
	// pre-render path templates
	for _, item := range c.Items {
		err := template.RenderField(&item.Source, &item.Path)
		if err != nil {
			return err
		}

		// remove and create destination
		if !store.DryRun {
			dest := utils.ExpandUser(item.Path)
			dest, _, _ = strings.Cut(dest, "/#/")
			if item.Replace {
				err := os.RemoveAll(dest)
				if err != nil {
					return err
				}
			}
			err := os.MkdirAll(dest, os.ModePerm)
			if err != nil {
				return err
			}
		}

	}
	output := log.NewMaxLineWriter(10)
	w, _ := f.(archiver.Walker)
	err = w.Walk(archive, func(f archiver.File) error {
		hName := getHeaderName(f)
		if hName == "" {
			return errors.New("invalid/unsupported archive file header")
		}
		for _, item := range c.Items {
			source := item.Source
			dest := utils.ExpandUser(item.Path)

			if matched, _ := doublestar.Match(source, hName); matched {
				// check if source is a wildcard path
				if isConstantMatch(source) {
					// exact path so handle new-name and use `dest` path
					// split <path> and <new-name>
					parts := strings.SplitN(dest, "/#/", 2)
					if len(parts) > 1 {
						dest = filepath.Join(parts[0], parts[1])
					} else {
						dest = filepath.Join(parts[0], filepath.Base(source))
					}
				} else {
					// wildcard path so handle strip components and use `dest` as the base path
					stripped := stripComponents(hName, item.Strip)
					if stripped == "" {
						return nil
					}
					dest = path.Join(dest, stripped)
				}

				if !store.DryRun {
					err := extractFile(f, dest)
					if err != nil {
						extractLogger.TagC(emerald.Red, "failed").Path(
							emerald.HighlightPath(hName, f.Mode()), emerald.HighlightPathStat(dest, nil),
						)
						return err
					}
				}

				fmt.Fprint(output, "[", emerald.Green, "+", emerald.Reset, "] ", emerald.HighlightPath(dest, f.Mode()), "\n")
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func stripComponents(path string, depth int) string {
	if strings.Count(path, "/") < depth {
		return ""
	}
	for i := 0; i < depth; i++ {
		slash := strings.Index(path, "/")
		path = path[slash+1:]
	}
	return path
}

func isConstantMatch(path string) bool {
	magicChars := `*?[`
	if runtime.GOOS != "windows" {
		magicChars = `*?[\`
	}
	return !strings.ContainsAny(path, magicChars)
}

func extractFile(f archiver.File, destination string) error {
	th, ok := f.Header.(*tar.Header)
	if ok {
		return untarFile(f, destination, th)
	}
	zfh, ok := f.Header.(zip.FileHeader)
	if ok {
		return unzipFile(f, destination, zfh)
	}
	return errors.New("unsupported archive header")
}

func untarFile(f archiver.File, destination string, hdr *tar.Header) error {
	switch hdr.Typeflag {
	case tar.TypeDir:
		return os.Mkdir(destination, f.Mode())
	case tar.TypeReg, tar.TypeChar, tar.TypeBlock, tar.TypeFifo, tar.TypeGNUSparse:
		return writeNewFile(destination, f, f.Mode())
	case tar.TypeXGlobalHeader:
		return nil // ignore the pax global header from git-generated tarballs
	default:
		return fmt.Errorf("%s: unknown type flag: %c", hdr.Name, hdr.Typeflag)
	}
}

func unzipFile(f archiver.File, destination string, hdr zip.FileHeader) error {
	if f.IsDir() || hdr.FileInfo().Mode()&os.ModeSymlink != 0 {
		return nil
	}
	return writeNewFile(destination, f, f.Mode())
}

func writeNewFile(fpath string, in io.Reader, fm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %v", fpath, err)
	}
	defer out.Close()

	err = out.Chmod(fm)
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("%s: changing file mode: %v", fpath, err)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %v", fpath, err)
	}
	return nil
}

func getHeaderName(f archiver.File) string {
	th, ok := f.Header.(*tar.Header)
	if ok {
		return th.Name
	}
	zfh, ok := f.Header.(zip.FileHeader)
	if ok {
		return zfh.Name
	}
	return ""
}
