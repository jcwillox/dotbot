package utils

import (
	"fmt"
	"github.com/jcwillox/dotbot/store"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"
)

type ExpandedPath string

func (path *ExpandedPath) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		fmt.Println(n.Value)
		fmt.Println(ExpandUser(n.Value))
		n.Value = ExpandUser(n.Value)
	}
	type ExpandedPathT ExpandedPath
	return n.Decode((*ExpandedPathT)(path))
}

func (path ExpandedPath) String() string {
	return string(path)
}

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
