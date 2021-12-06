package utils

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type WeakFileMode os.FileMode

func (w *WeakFileMode) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" {
		fmt.Println("mode", *w)
		*w = WeakFileMode(FileModeFromString(n.Value, os.FileMode(*w)))
		return nil
	}
	type WeakFileModeT WeakFileMode
	return n.Decode((*WeakFileModeT)(w))
}

func FileModeFromString(mode string, mask os.FileMode) os.FileMode {
	if mode == "+x" {
		return mask | 0111
	}
	return mask
}
