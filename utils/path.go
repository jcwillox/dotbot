package utils

import (
	"github.com/jcwillox/dotbot/template"
	"golang.org/x/sys/execabs"
	"log"
	"os"
	"strings"
)

func PathHasExecutable(file string) bool {
	path, _ := execabs.LookPath(file)
	return path != ""
}

// StripPath removes all specified paths from the PATH environment variable on WSL
// this provides a significant performance improvement as it normally includes many
// large networked windows directories which are very slow to access
func StripPath(paths ...string) {
	if paths == nil || !template.IsWSL() {
		return
	}
	envPath := os.Getenv("PATH")
	newPath := make([]string, 0, 10)
	skip := false
	for _, line := range strings.Split(envPath, string(os.PathListSeparator)) {
		skip = false
		for _, path := range paths {
			if path != "" && strings.HasPrefix(line, path) {
				skip = true
				break
			}
		}
		if !skip {
			newPath = append(newPath, line)
		}
	}
	err := os.Setenv("PATH", strings.Join(newPath, string(os.PathListSeparator)))
	if err != nil {
		log.Panicln("Failed to set PATH env var with stripped path", err)
	}
}
