package template

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"text/template"
)

var funcs = map[string]interface{}{
	"env":         os.Getenv,
	"Distro":      Distro,
	"MatchDistro": MatchDistro,
	"OS":          func() string { return runtime.GOOS },
	"ARCH":        func() string { return runtime.GOARCH },
	"IsWSL":       IsWSL,
}

var tmplVars = make(map[string]interface{})

func Vars(vars map[string]interface{}) {
	for key, newVal := range vars {
		tmplVars[key] = newVal
	}
}

func GetVar(key string) (value interface{}, present bool) {
	value, present = tmplVars[key]
	return value, present
}

func GetVars() map[string]interface{} {
	return tmplVars
}

func VarsClosure(vars map[string]interface{}) func() {
	prev := make(map[string]interface{})
	for key, newVal := range vars {
		if val, present := tmplVars[key]; present {
			prev[key] = val
		}
		tmplVars[key] = newVal
	}
	return func() {
		// iterate over changed keys and restore old value
		for key := range vars {
			if val, present := prev[key]; present {
				tmplVars[key] = val
			} else {
				// remove key if no old value
				delete(tmplVars, key)
			}
		}
	}
}

type Template struct {
	Template *template.Template
}

func (t *Template) Parse(tmpl string) *Template {
	t.Template = template.Must(t.Template.Funcs(funcs).Parse(tmpl))
	return t
}

func New(name string) *Template {
	return &Template{template.New(name)}
}

func Parse(tmpl string) *Template {
	return New("").Parse(tmpl)
}

func (t *Template) Funcs(funcMap template.FuncMap) *Template {
	t.Template = t.Template.Funcs(funcMap)
	return t
}

func Funcs(funcMap template.FuncMap) *Template {
	return New("").Funcs(funcMap)
}

func (t *Template) Render() (string, error) {
	var buff bytes.Buffer
	err := t.Template.Execute(&buff, tmplVars)
	return buff.String(), err
}

// HasTemplate returns true if the string has '{{' followed by '}}'
func HasTemplate(tmpl string) bool {
	first := strings.Index(tmpl, "{{")
	if first < 0 {
		return false
	}
	last := strings.Index(tmpl, "}}")
	if last > 0 && last > first {
		return true
	}
	return false
}
