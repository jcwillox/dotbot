package utils

import (
	"bufio"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
	"golang.org/x/sys/execabs"
	"os"
	"path/filepath"
	"strings"
)

func ExpandUser(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return path
	}
	return filepath.Join(store.HomeDirectory, path[1:])
}

func ShrinkUser(path string) string {
	if !strings.HasPrefix(path, store.HomeDirectory) {
		return path
	}
	length := len(store.HomeDirectory)
	if len(path) > length && path[length] != '/' && path[length] != '\\' {
		return path
	}
	return filepath.Join("~", path[length:])
}

func GetConfigPath() string {
	if v, present := os.LookupEnv("DOTBOT_CONFIG"); present {
		return v
	}
	for _, ext := range []string{"yaml", "yml", "json"} {
		filename := "dotbot." + ext
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}
	return ""
}

func EnsureInBaseDir() {
	if base, present := store.HasGet("directory"); present {
		err := os.Chdir(base)
		if err != nil {
			log.Fatalln("Unable to access dotfiles directory", err)
		}
	}
}

var isMusl = -1

func IsMusl() bool {
	if isMusl > -1 {
		return isMusl == 0
	}
	isLibcMusl, err := isLibcMusl()
	if err != nil {
		log.Fatalln("failed detecting system libc", err)
	}
	if isLibcMusl {
		isMusl = 0
		return true
	} else {
		isMusl = 1
		return false
	}
}

func isLibcMusl() (bool, error) {
	// perform quick file checks
	if _, err := os.Stat("/lib/ld-musl-x86_64.so.1"); err == nil {
		return true, nil
	}
	if _, err := os.Stat("/lib64/ld-linux-x86-64.so.2"); err == nil {
		return false, nil
	}
	// fallback to checking ldd
	cmd := execabs.Command("ldd", "--version")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}
	reader := bufio.NewReader(stdout)
	err = cmd.Start()
	if err != nil {
		return false, err
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	if strings.HasPrefix(line, "musl") {
		return true, nil
	} else {
		return false, nil
	}
}
