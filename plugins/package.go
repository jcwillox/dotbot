package plugins

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/utils/sudo"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"golang.org/x/sys/execabs"
	"gopkg.in/yaml.v3"
)

var packageLogger = log.NewBasicLogger("PACKAGE")

type PackageBase []PackageConfig
type PackageConfig []*PackageItem
type PackageItem struct {
	Manager  string `yaml:",omitempty"`
	Packages []string
}

func (b *PackageBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type PackageBaseT PackageBase
	return n.Decode((*PackageBaseT)(b))
}

func (p *PackageConfig) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.ScalarToMapVal(n, "os")
	n = yamltools.MapToSliceMap(n)
	type PackageConfigT PackageConfig
	return n.Decode((*PackageConfigT)(p))
}
func (c *PackageItem) UnmarshalYAML(n *yaml.Node) error {
	n.Content[1] = yamltools.EnsureList(n.Content[1])
	n = yamltools.MapSplitKeyVal(n, "manager", "packages")
	type PackageItemT PackageItem
	return n.Decode((*PackageItemT)(c))
}

func (c *PackageItem) MarshalYAML() (interface{}, error) {
	manager := c.Manager
	c.Manager = ""
	type PackageItemT PackageItem
	return map[string][]string{manager: c.Packages}, nil
}

func (b PackageBase) Enabled() bool {
	return sudo.CanSudo() || sudo.IsRoot()
}

func (b PackageBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

func (p PackageConfig) Run() error {
	for _, c := range p {
		if c.Manager == "os" {
			c.Manager = getOsPackager()
		}
		if utils.OnPath(c.Manager) {
			return c.InstallAll()
		}
	}
	return nil
}

func (c PackageItem) InstallAll() error {
	var command *utils.Command
	switch c.Manager {
	case "apt":
		for _, pkg := range c.Packages {
			version, latest := getAptPackageVersion(pkg)
			logPackage(pkg, version, latest)
			if version == latest || store.DryRun {
				break
			}
			command = &utils.Command{
				Command:  "apt-get install -qq -y " + pkg,
				Shell:    false,
				Stdout:   false,
				Stderr:   true,
				Sudo:     true,
				MaxLines: 10,
			}

		}
	case "apk":
		for _, pkg := range c.Packages {
			version, latest := getApkPackageVersion(pkg)
			logPackage(pkg, version, latest)
			if version == latest || store.DryRun {
				break
			}
			command = &utils.Command{
				Command:  "apk add " + pkg,
				Shell:    false,
				Stdout:   true,
				Stderr:   true,
				Sudo:     true,
				MaxLines: 10,
			}
		}
	case "brew":
		for _, pkg := range c.Packages {
			version, latest := getBrewPackageVersion(pkg)
			logPackage(pkg, version, latest)
			if version == latest || store.DryRun {
				break
			}
			command = &utils.Command{
				Command:  "brew install -q " + pkg,
				Shell:    false,
				Stdout:   true,
				Stderr:   true,
				Sudo:     false,
				MaxLines: 10,
			}
		}
	}
	if command == nil {
		return nil
	}
	cmd, err := command.Cmd()
	if err != nil {
		return err
	}
	return cmd.Run()
}

var highlightVersion = emerald.ColorFunc("cyan+u")

func logPackage(pkg string, version string, latest string) {
	if version == "" && latest == "" {
		packageLogger.TagC(emerald.Red, "invalid").Print(emerald.Green, pkg, "\n")
	} else if version == latest {
		packageLogger.TagDone("up-to-date").Print(
			emerald.Green, pkg, " ", emerald.Blue, version, "\n",
		)
	} else if version == "" {
		packageLogger.TagSudo("installing", true).Print(emerald.Green, pkg, " ", highlightVersion(latest), "\n")
	} else {
		packageLogger.TagSudo("updating", true).Print(
			emerald.Green, pkg, emerald.Reset, " ", emerald.Blue, version,
			emerald.LightBlack, " -> ", highlightVersion(latest), "\n",
		)
	}
}

func getOsPackager() string {
	switch {
	// MacOS has a binary called apt, but it's not the package manager, so brew needs to come first
	case utils.OnPath("brew"):
		return "brew"
	case utils.OnPath("apt"):
		return "apt"
	case utils.OnPath("apk"):
		return "apk"
	default:
		return ""
	}
}

func getAptPackageVersion(pkg string) (installed string, latest string) {
	cmd := execabs.Command("apt-cache", "policy", pkg)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("Installed:")) {
			line = line[11:]
			if bytes.Equal(line, []byte("(none)")) {
				installed = ""
			} else {
				installed = string(line)
			}
		} else if bytes.HasPrefix(line, []byte("Candidate:")) {
			line = line[11:]
			if bytes.Equal(line, []byte("(none)")) {
				latest = ""
			} else {
				latest = string(line)
			}
		}
	}
}

func getApkPackageVersion(pkg string) (installed string, latest string) {
	cmd := execabs.Command("apk", "policy", pkg)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		if line[0] != ' ' {
			bytes.TrimSuffix(bytes.TrimSpace(line), []byte(" policy:"))
			version, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			version = bytes.TrimSpace(version)
			repo, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			if bytes.Equal(bytes.TrimSpace(repo), []byte("lib/apk/db/installed")) {
				installed = string(version[:len(version)-1])
			} else {
				latest = string(version[:len(version)-1])
			}
		}
		if bytes.HasPrefix(line, []byte("   ")) {
			continue
		}
		if bytes.HasPrefix(line, []byte("  ")) {
			line = bytes.TrimSpace(line)
			latest = string(line[:len(line)-1])
			break
		}
	}
	if installed != "" && latest == "" {
		latest = installed
	}
	return installed, latest
}

func getBrewPackageVersion(pkg string) (installed string, latest string) {
	cmd := execabs.Command("brew", "info", "--json=v2", pkg)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	stdoutContent, err := io.ReadAll(stdout)
	if err != nil {
		return
	}

	var info struct {
		Formulae []struct {
			Name     string `json:"full_name"`
			Outdated bool   `json:"outdated"`
			Versions struct {
				Stable string `json:"stable"`
			} `json:"versions"`
			Installed []struct {
				Version string `json:"version"`
			} `json:"installed"`
		} `json:"formulae"`
		Casks []struct {
			Token     string `json:"token"`
			Version   string `json:"version"`
			Installed string `json:"installed"`
		} `json:"casks"`
	}
	err = json.Unmarshal(stdoutContent, &info)
	if err != nil {
		return
	}

	if len(info.Formulae) > 0 {
		for _, f := range info.Formulae {
			if f.Name != pkg {
				continue
			}
			if len(f.Installed) < 1 {
				return "", f.Versions.Stable
			}
			// Some installed versions do not match the latest versions
			// e.g. if a new formula version was released for the same application version
			// Just return the installed version as the latest version if the formula is not outdated
			if !f.Outdated {
				return f.Installed[0].Version, f.Installed[0].Version
			}
			return f.Installed[0].Version, f.Versions.Stable
		}
	}

	if len(info.Casks) > 0 {
		for _, c := range info.Casks {
			if c.Token != pkg {
				continue
			}
			return c.Installed, c.Version
		}
	}

	return
}
