package plugins

import (
	"fmt"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/utils/sudo"
	"github.com/jcwillox/dotbot/yamltools"
	"github.com/shirou/gopsutil/host"
	"gopkg.in/yaml.v3"
	"runtime"
)

type SharkdpBase []SharkdpConfig
type SharkdpConfig string

func (b *SharkdpBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.EnsureList(n)
	type SharkdpBaseT SharkdpBase
	return n.Decode((*SharkdpBaseT)(b))
}

func (b SharkdpBase) Enabled() bool {
	return true
}

func (b SharkdpBase) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

func (c SharkdpConfig) Run() error {
	name := string(c)
	url := "https://github.com/sharkdp/" + name
	_, family, _, err := host.PlatformInformation()
	if err != nil {
		return err
	}
	if family == "debian" && sudo.CanSudo() {
		return InstallConfig{
			Name: name,
			Url:  url,
			Sudo: true,
			Download: &DownloadConfig{
				Url:  "/releases/download/v{{ .Version }}/" + name + "_{{ .Version }}_{{ ARCH }}.deb",
				Mode: 438,
			},
			Shell: &ShellConfig{
				Command: utils.Command{
					Command: "dpkg -i {{ .Path }}",
				},
			},
		}.Run()
	} else if runtime.GOOS == "linux" {
		asset := name + "-v{{ .Version }}"
		if runtime.GOARCH == "amd64" {
			asset += "-x86_64"
		} else if runtime.GOARCH == "386" {
			asset += "-i686"
		} else if runtime.GOARCH == "arm64" {
			asset += "-aarch64"
		} else {
			asset += "-" + runtime.GOARCH
		}
		asset += "-unknown-linux"
		if utils.IsMusl() {
			asset += "-musl"
		} else {
			asset += "-gnu"
		}

		var items ExtractItems
		if sudo.CanSudo() {
			items = ExtractItems{
				{
					Source: asset + "/" + name,
					Path:   "/usr/local/bin",
				},
				{
					Source: asset + "/autocomplete/" + name + ".zsh",
					Path:   "/usr/local/share/zsh/site-functions/#/_" + name,
				},
				{
					Source: asset + "/autocomplete/_" + name,
					Path:   "/usr/local/share/zsh/site-functions",
				},
				{
					Source: asset + "/autocomplete/" + name + ".fish",
					Path:   "/usr/share/fish/completions",
				},
				{
					Source: asset + "/autocomplete/" + name + ".bash",
					Path:   "/etc/bash_completion.d/#/" + name,
				},
				{
					Source: asset + "/" + name + ".1",
					Path:   "/usr/local/share/man/man1",
				},
			}
		} else {
			items = ExtractItems{
				{
					Source: asset + "/" + name,
					Path:   "~/.local/bin",
				},
				{
					Source: asset + "/" + name + ".1",
					Path:   "~/.local/share/man/man1",
				},
			}
		}

		return InstallConfig{
			Name: name,
			Url:  url,
			Sudo: sudo.CanSudo(),
			Download: &DownloadConfig{
				Url:     "/releases/download/v{{ .Version }}/" + asset + ".tar.gz",
				Extract: items,
				Mode:    438,
			},
		}.Run()
	}
	return nil
}
