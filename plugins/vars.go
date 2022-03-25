package plugins

import (
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/template"
	"github.com/jcwillox/dotbot/yamltools"
	"gopkg.in/yaml.v3"
)

type VarsBase []map[string]interface{}

func (b *VarsBase) UnmarshalYAML(n *yaml.Node) error {
	n = yamltools.MapToSliceMap(n)
	n = yamltools.EnsureList(n)
	type VarsBaseT VarsBase
	return n.Decode((*VarsBaseT)(b))
}

func (b VarsBase) Enabled() bool {
	return true
}

func (b VarsBase) RunAll() error {
	for _, config := range b {
		for k, v := range config {
			if s, ok := v.(string); ok {
				err := template.RenderField(&s)
				if err != nil {
					return err
				}
				store.TmplVar(k, s)
			} else {
				store.TmplVar(k, v)
			}
		}
	}
	return nil
}
