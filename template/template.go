package template

import (
	"bytes"
	"github.com/jcwillox/dotbot/utils"
	"os"
	"runtime"
	"text/template"
)

var funcs = map[string]interface{}{
	"env":         os.Getenv,
	"Distro":      Distro,
	"MatchDistro": MatchDistro,
	"OS":          func() string { return runtime.GOOS },
	"ARCH":        func() string { return runtime.GOARCH },
	"IsWSL":       utils.IsWSL,
}

func RenderTemplate(tmpl string) string {
	var buff bytes.Buffer
	err := template.Must(
		template.New("").Funcs(funcs).Parse(tmpl),
	).Execute(&buff, "no data needed")
	if err != nil {
		panic(err)
	}
	result := buff.String()
	return result
}
