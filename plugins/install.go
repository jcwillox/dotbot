package plugins

import (
	"errors"
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/utils/sudo"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
)

var installLogger = log.NewBasicLogger("INSTALL")

type InstallBase []InstallConfig
type InstallConfig struct {
	Name     string
	Url      string
	Version  string
	Download *DownloadConfig
	Shell    *ShellConfig
	Sudo     bool
	TrySudo  bool `yaml:"try_sudo"`
	Then     PluginList
}

func (b *InstallBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type InstallBaseT InstallBase
	return n.Decode((*InstallBaseT)(b))
}

func (b InstallBase) Enabled() bool {
	return true
}

func (b InstallBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

func logInstall(title string, version string, latest string) {
	if version == "" && latest == "" {
		installLogger.TagC(emerald.Red, "invalid").Print(emerald.Green, title, "\n")
	} else if version == latest {
		installLogger.TagDone("up-to-date").Print(
			emerald.Green, title, " ", emerald.Blue, version, "\n",
		)
	} else if version == "" {
		installLogger.Tag("installing").Print(emerald.Green, title, " ", highlightVersion(latest), "\n")
	} else {
		installLogger.Tag("updating").Print(
			emerald.Green, title, emerald.Reset, " ", emerald.Blue, version,
			emerald.LightBlack, " -> ", highlightVersion(latest), "\n",
		)
	}
}

func (c InstallConfig) Run() error {
	var version string
	var err error
	if strings.HasPrefix(c.Url, "https://github.com/") && c.Version == "" {
		version, err = GetGithubVersion(c.Url)
	} else {
		version, err = GetGenericVersion(c.Url, c.Version)
	}
	if err != nil {
		return err
	}
	if version == "" {
		return errors.New("latest version was empty")
	}

	current := store.Get(c.Url)

	// abort early if we don't have root privileges
	if c.Sudo && !sudo.CanSudo() {
		return nil
	}

	logInstall(c.String(), current, version)
	if current != version {
		defer store.VarsClosure(map[string]interface{}{"Current": current, "Version": version, "Url": c.Url})()

		// merge shorthand directives into then block
		then := make(PluginList, 0, 2)
		if c.Download != nil {
			then = append(then, map[string]Plugin{"download": DownloadBase{c.Download}})
		}
		if c.Shell != nil {
			then = append(then, map[string]Plugin{"shell": ShellBase{c.Shell}})
		}
		if len(then) > 0 {
			c.Then = append(then, c.Then...)
		}

		if c.Sudo || c.TrySudo {
			if sudo.WouldSudo() {
				err := sudo.Configs(&c.Then)
				if err != nil {
					return err
				}
			} else if sudo.IsRoot() || c.TrySudo {
				c.Then.RunAll()
			}
		} else {
			c.Then.RunAll()
		}

		if !store.DryRun {
			store.SetSave(c.Url, version)
		}
	}
	return nil
}

func (c InstallConfig) String() string {
	if c.Name != "" {
		return c.Name
	}
	return c.Url
}

var noFollowClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func GetGithubVersion(url string) (string, error) {
	if !strings.HasSuffix(url, "/releases/latest") {
		url = strings.TrimRight(url, "/") + "/releases/latest"
	}
	resp, err := noFollowClient.Head(url)
	if err != nil {
		return "", err
	}
	location, err := resp.Location()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(path.Base(location.Path), "v"), nil
}

func GetGenericVersion(url, versionTmpl string) (string, error) {
	if template.HasTemplate(versionTmpl) {
		return template.Parse(versionTmpl).Render()
	} else {
		versionRegex, err := regexp.Compile(versionTmpl)
		if err != nil {
			return "", err
		}
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		log.Debugln("fetching", url)
		matches := versionRegex.FindSubmatch(data)
		for i, match := range matches {
			log.Debugf("[match] %d: %s\n", i, match)
		}
		if len(matches) > 1 {
			return string(matches[1]), nil
		} else if matches == nil {
			return "", nil
		}
		return string(matches[0]), nil
	}
}
