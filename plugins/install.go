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
	Version  InstallVersion
	Download *DownloadConfig
	Shell    *ShellConfig
	Sudo     bool
	TrySudo  bool `yaml:"try_sudo"`
	Then     PluginList
}
type InstallVersion struct {
	Url   string `yaml:",omitempty"`
	Regex string
}

func (b *InstallBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type InstallBaseT InstallBase
	return n.Decode((*InstallBaseT)(b))
}

func (c *InstallVersion) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.ScalarToMapVal(n, "regex")
	type VersionConfigT InstallVersion
	return n.Decode((*VersionConfigT)(c))
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
	version, err := GetVersion(c.Url, &c.Version)
	if err != nil {
		return err
	}
	if version == "" {
		return errors.New("latest version was empty")
	}
	current := store.Get(c.Version.Url)

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
			store.SetSave(c.Version.Url, version)
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

func GetVersion(baseUrl string, config *InstallVersion) (string, error) {
	if strings.HasPrefix(config.Url, "/") || config.Url == "" {
		config.Url = baseUrl + config.Url
	}
	if config.Regex != "" {
		return GetRegexVersion(*config)
	} else if strings.HasPrefix(config.Url, "https://github.com/") {
		return GetGithubVersion(config.Url)
	}
	return "", errors.New("could not determine method to extract version")
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

func GetRegexVersion(c InstallVersion) (string, error) {
	if template.HasTemplate(c.Regex) {
		return template.Parse(c.Regex).Render()
	} else {
		versionRegex, err := regexp.Compile(c.Regex)
		if err != nil {
			return "", err
		}
		resp, err := http.Get(c.Url)
		if err != nil {
			return "", err
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		log.Debugln("fetching", c.Url)
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
