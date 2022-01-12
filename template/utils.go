package template

import (
	"bufio"
	"golang.org/x/sys/execabs"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func IsWSL() bool {
	_, isWSL := os.LookupEnv("WSL_DISTRO_NAME")
	return isWSL
}

func DefaultShell() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	uid := strconv.Itoa(os.Getuid())
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return ""
		}
		parts := strings.Split(strings.TrimSpace(line), ":")
		if len(parts) < 7 {
			continue
		}
		if parts[2] == uid {
			return parts[6]
		}
	}
}

func Which(file string) string {
	path, _ := execabs.LookPath(file)
	return path
}
