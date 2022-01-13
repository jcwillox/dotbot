//go:build !linux
// +build !linux

package template

import (
	"github.com/shirou/gopsutil/host"
	"strings"
)

func getDistro() string {
	platform, _, _, err := host.PlatformInformation()
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(platform, "Microsoft ")
}
