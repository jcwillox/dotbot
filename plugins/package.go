package plugins

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/jcwillox/emerald"
	"golang.org/x/sys/execabs"
	"gopkg.in/yaml.v3"
)

type PackageBase []PackageConfig
type PackageConfig []Package
type Package struct {
	Manager  string
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
func (c *Package) UnmarshalYAML(n *yaml.Node) error {
	n.Content[1] = yamltools.EnsureList(n.Content[1])
	n = yamltools.MapSplitKeyVal(n, "manager", "packages")
	type PackageT Package
	return n.Decode((*PackageT)(c))
}

func (b PackageBase) Enabled() bool {
	return true
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

var packageLogger = log.GetLogger(emerald.ColorCode("blue+b"), "PACKAGE", emerald.Yellow)

func (p PackageConfig) Run() error {
	for _, c := range p {
		if c.Manager == "os" {
			c.Manager = getOsPackager()
		}
		if utils.PathHasExecutable(c.Manager) {
			return c.InstallAll()
		}
	}
	return nil
}

func (c Package) InstallAll() error {
	switch c.Manager {
	case "apt":
		for _, pkg := range c.Packages {
			version, latest := getAptPackageVersion(pkg)
			logPackage(pkg, version, latest)
			if version == latest {
				break
			}
			err := utils.Command{
				Command: "apt-get install -qq -y " + pkg,
				Shell:   false,
				Stdout:  true,
				Stderr:  true,
				Sudo:    true,
			}.Run()
			if err != nil {
				return err
			}
		}
	case "apk":
		for _, pkg := range c.Packages {
			version, latest := getApkPackageVersion(pkg)
			logPackage(pkg, version, latest)
			if version == latest {
				break
			}
			err := utils.Command{
				Command: "apk add " + pkg,
				Shell:   false,
				Stdout:  true,
				Stderr:  true,
				Sudo:    true,
			}.Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var highlightVersion = emerald.ColorFunc("cyan+u")

func logPackage(pkg string, version string, latest string) {
	if version == "" && latest == "" {
		packageLogger.Log().TagC(emerald.Red, "invalid").Print(emerald.Green, pkg, "\n")
	} else if version == latest {
		packageLogger.Log().TagC(emerald.LightBlack, "up-to-date").Print(
			emerald.Green, pkg, " ", emerald.Blue, version, "\n",
		)
	} else if version == "" {
		packageLogger.Log().Tag("installing").Print(emerald.Green, pkg, " ", highlightVersion(latest), "\n")
	} else {
		packageLogger.Log().Tag("updating").Print(
			emerald.Green, pkg, emerald.Reset, " ", emerald.Blue, version,
			emerald.LightBlack, " -> ", highlightVersion(latest), "\n",
		)
	}
}

func getOsPackager() string {
	switch {
	case utils.PathHasExecutable("apt"):
		return "apt"
	case utils.PathHasExecutable("apk"):
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
