package utils

import (
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/dotbot/store"
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
