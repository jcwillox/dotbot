//go:build ignore
// +build ignore

package plugin

import (
	"fmt"
	"github.com/jcwillox/dotbot/log"
	"github.com/jcwillox/emerald"
	"gopkg.in/yaml.v3"
)

type Base []Config

type Config struct {
}

func (b *Base) UnmarshalYAML(n *yaml.Node) error {
	type BaseT Base
	return n.Decode((*BaseT)(b))
}

func (b Base) Enabled() bool {
	return true
}

func (b Base) RunAll() error {
	for _, config := range b {
		err := config.Run()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
	return nil
}

var logger = log.GetLogger(emerald.White, "PLUGIN", emerald.Blue)

func (c Config) Run() error {
	return nil
}
