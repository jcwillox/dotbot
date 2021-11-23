package plugins

import (
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/yamltools"
	"gopkg.in/yaml.v3"
)

type StripPathBase []string

func (b *StripPathBase) UnmarshalYAML(n *yaml.Node) error {
	if val, ok := yamltools.ParseBoolNode(n); ok {
		if val {
			*b = []string{"/mnt/c"}
		} else {
			*b = []string{""}
		}
		return nil
	} else {
		n = yamltools.EnsureList(n)
		type StripPathBaseT StripPathBase
		return n.Decode((*StripPathBaseT)(b))
	}
}

func (b StripPathBase) Run() {
	utils.StripPath(b...)
}
