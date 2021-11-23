package utils

import (
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

func IsWSL() bool {
	_, isWSL := os.LookupEnv("WSL_DISTRO_NAME")
	return isWSL
}
