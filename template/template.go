package template

import (
	"bytes"
	"github.com/jcwillox/dotbot/store"
	"github.com/jcwillox/dotbot/utils"
	"github.com/jcwillox/dotbot/utils/sudo"
	"golang.org/x/sys/execabs"
	"os"
	"runtime"
	"strings"
	"text/template"
)

var funcs = map[string]interface{}{
	"env":          os.Getenv,
	"Distro":       Distro,
	"MatchDistro":  MatchDistro,
	"OS":           func() string { return runtime.GOOS },
	"ARCH":         func() string { return runtime.GOARCH },
	"DefaultShell": utils.DefaultShell,
	"IsWSL":        utils.IsWSL,
	"IsMusl":       utils.IsMusl,
	"LIBC":         utils.GetLibc,
	"IsRoot":       sudo.IsRoot,
	"CanSudo":      sudo.CanSudo,
	"OnPath":       utils.OnPath,
	"Which": func(file string) string {
		path, _ := execabs.LookPath(file)
		return path
	},
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
	err := t.Template.Execute(&buff, store.GetVars())
	return buff.String(), err
}

func (t *Template) RenderTrue() (bool, error) {
	result, err := t.Render()
	if err != nil {
		return false, err
	}
	return strings.EqualFold(result, "true"), nil
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

func renderField(tmpl *string) error {
	if !HasTemplate(*tmpl) {
		return nil
	}
	result, err := Parse(*tmpl).Render()
	if err != nil {
		return err
	}
	*tmpl = result
	return nil
}

func RenderField(tmpls ...*string) error {
	for _, tmpl := range tmpls {
		err := renderField(tmpl)
		if err != nil {
			return err
		}
	}
	return nil
}
