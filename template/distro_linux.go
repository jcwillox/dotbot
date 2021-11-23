//go:build !windows
// +build !windows

package template

import (
	"bufio"
	"bytes"
	"golang.org/x/sys/execabs"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getPrettyName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue
		}
		if parts[0] == "PRETTY_NAME" || parts[0] == "DISTRIB_DESCRIPTION" {
			return strings.Trim(parts[1], "\"\n")
		}
	}
	return ""
}

func getProxmoxVersion() string {
	if _, err := exec.LookPath("pveversion"); err != nil {
		return ""
	}
	cmd := execabs.Command("dpkg", "-s", "pve-manager")
	stdout, err := cmd.StdoutPipe()
	if err == nil {
		reader := bufio.NewReader(stdout)
		err = cmd.Start()
		if err == nil {
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				if strings.HasPrefix(line, "Version:") {
					return "Proxmox VE " + strings.TrimRight(strings.TrimPrefix(line, "Version: "), "\n")
				}
			}
		}
	}
	return "Proxmox VE"
}

func getDistro() string {

	name := getProxmoxVersion()
	if name != "" {
		return name
	}
	name = getPrettyName("/etc/lsb-release")
	if name != "" {
		return name
	}
	out, err := execabs.Command("lsb_release", "-sd").Output()
	if err == nil {
		return string(bytes.TrimRight(out, "\n"))
	}
	name = getPrettyName("/etc/os-release")
	if name != "" {
		return name
	}
	name = getPrettyName("/usr/lib/os-release")
	if name != "" {
		return name
	}
	name = getPrettyName("/etc/openwrt_release")
	if name != "" {
		return name
	}
	out, err = execabs.Command("uname", "-o").Output()
	if err == nil {
		return string(bytes.TrimRight(out, "\n"))
	}
	return runtime.GOOS
}
