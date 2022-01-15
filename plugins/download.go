package plugins

import (
	"fmt"
	"github.com/creasty/defaults"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/utils/sudo"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DownloadBase []*DownloadConfig
type DownloadConfig struct {
	Name    string
	Url     string
	Path    string `yaml:",omitempty"`
	Mkdirs  bool   `default:"true"`
	Force   bool
	Mode    utils.WeakFileMode `default:"438"`
	Extract ExtractItems
}

func (b *DownloadBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type DownloadBaseT DownloadBase
	return n.Decode((*DownloadBaseT)(b))
}

func (c *DownloadConfig) UnmarshalYAML(n *yaml.Node) error {
	defaults.MustSet(c)
	n = yamltools.MapKeyIntoValueMap(n, "path")
	type DownloadConfigT DownloadConfig
	return n.Decode((*DownloadConfigT)(c))
}

func (c *DownloadConfig) MarshalYAML() (interface{}, error) {
	path := c.Path
	c.Path = ""
	type DownloadConfigT DownloadConfig
	return map[string]*DownloadConfigT{path: (*DownloadConfigT)(c)}, nil
}

func (b DownloadBase) Enabled() bool {
	return true
}

func (b DownloadBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if sudo.IsPermission(err) && sudo.WouldSudo() {
			if !sudo.HasUsedSudo {
				// let user know why we want to sudo
				downloadLogger.Log().TagC(emerald.Yellow, "downloading").Sudo(true).Print(
					emerald.HighlightFileMode(os.FileMode(config.Mode)), " ", emerald.HighlightPath(config.Path, os.FileMode(config.Mode)), "\n",
				)
			}
			err = sudo.Config("download", &config)
		}
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
	}
	return nil
}

var downloadLogger = log.GetLogger(emerald.ColorCode("blue+b"), "DOWNLOAD", emerald.Yellow)

func (c *DownloadConfig) Run() error {
	var f *os.File
	// allow templating
	err := template.RenderField(&c.Url, &c.Path)
	if err != nil {
		return err
	}
	// special handling of urls starting with '/'
	if strings.HasPrefix(c.Url, "/") {
		if url, present := store.GetVar("Url"); present {
			c.Url = url.(string) + c.Url
		}
	}
	// grab filename from url
	name := filepath.Base(c.Url)
	if c.Path == "" {
		// use proper download file does not exist
		path := filepath.Join(os.TempDir(), name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(c.Mode))
			if err != nil {
				return err
			}
		} else {
			ext := filepath.Ext(name)
			f, err = os.CreateTemp("", "dotbot-*"+ext)
			if err != nil {
				return err
			}
		}
		// track temp file for deletion
		store.TrackTempFile(f.Name())
		store.Vars(map[string]interface{}{"Path": f.Name()})
	} else {
		// extract filename
		path := utils.ExpandUser(c.Path)
		if !strings.HasSuffix(path, name) {
			if stat, err := os.Stat(path); err == nil {
				if stat.IsDir() {
					// path is existing directory, append remote name to path
					path = filepath.Join(path, filepath.Base(c.Url))
				} else if !c.Force {
					// skip as file is already present and force is not set
					return nil
				}
			}
		}
		if _, err := os.Stat(path); err == nil && !c.Force {
			// skip as file is already present and force is not set
			return nil
		}
		name = filepath.Base(path)
		if c.Mkdirs {
			err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return err
			}
		}
		f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(c.Mode))
		if err != nil {
			return err
		}
	}

	// get actual download length
	head, err := http.Head(c.Url)
	if err != nil {
		return err
	}

	// download file
	resp, err := http.Get(c.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if c.Name != "" {
		name = c.Name
	}

	// download file using progress bar when color enabled
	if emerald.ColorEnabled {
		emerald.HideCursor()
		defer emerald.ShowCursor()

		p := mpb.New(
			mpb.WithRefreshRate(100 * time.Millisecond),
		)

		bar := AddProgressBar(p, head.ContentLength, emerald.Bold+emerald.Blue+name+emerald.Reset)
		proxyReader := bar.ProxyReader(resp.Body)
		defer proxyReader.Close()

		_, err = io.Copy(f, proxyReader)
		if err != nil {
			return err
		}

		p.Wait()
	} else {
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
	}
	err = f.Close()
	if err != nil {
		return err
	}
	if c.Extract != nil && len(c.Extract) > 0 {
		return ExtractConfig{
			Archive: f.Name(),
			Items:   c.Extract,
		}.Run()
	}
	return nil
}

func AddProgressBar(p *mpb.Progress, total int64, desc string) *mpb.Bar {
	if total < 0 {
		return p.Add(total,
			mpb.NewBarFiller(mpb.SpinnerStyle()),
			mpb.BarWidth(1),
			mpb.BarFillerClearOnComplete(),
			mpb.BarFillerTrim(),
			mpb.PrependDecorators(
				decor.Name(emerald.ColorIndexFg(161)),
				decor.OnComplete(decor.Name(" "), ""),
			),
			mpb.AppendDecorators(
				decor.OnComplete(decor.Name(emerald.Reset+" • "), emerald.Reset),
				decor.Name(desc),
				decor.Name(emerald.Reset+" • "+emerald.Green),
				decor.CurrentKiloByte("% .1f"),
				decor.Name(emerald.Reset+" • "+emerald.Red),
				decor.EwmaSpeed(decor.UnitKB, "% .2f", 160),
				decor.Name(emerald.Reset+" • "+emerald.Cyan),
				decor.Elapsed(decor.ET_STYLE_MMSS),
				decor.Name(emerald.Reset, decor.WCSyncWidth),
			),
		)
	} else {
		return p.Add(total,
			mpb.NewBarFiller(mpb.BarStyle().Lbound(emerald.ColorIndexFg(161)).Filler("━").Padding(emerald.LightBlack+"━").Tip(emerald.ColorIndexFg(161)+"╸").Rbound(emerald.Reset)),
			mpb.PrependDecorators(
				decor.Name(desc),
			),
			mpb.AppendDecorators(
				decor.Name(emerald.Magenta),
				decor.NewPercentage("%.1f"),
				decor.Name(emerald.Reset+" • "+emerald.Green),
				decor.OnComplete(decor.CurrentKiloByte("% .1f"), ""),
				decor.OnComplete(decor.Name("/"), ""),
				decor.TotalKiloByte("% .1f"),
				decor.Name(emerald.Reset+" • "+emerald.Red),
				decor.EwmaSpeed(decor.UnitKB, "% .2f", 160),
				decor.Name(emerald.Reset+" • "+emerald.Cyan),
				decor.Elapsed(decor.ET_STYLE_MMSS),
				decor.OnComplete(decor.Name("<"), ""),
				decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 160), ""),
				decor.Name(emerald.Reset),
			),
		)
	}
}
