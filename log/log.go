package log

import (
	"github.com/jcwillox/emerald"
)

func GetLogger(color string, name string, tagColor string) Logger {
	return Logger{
		name:     name,
		color:    color,
		tagColor: tagColor,
	}
}

type Logger struct {
	name     string
	color    string
	tagColor string
}

func (l Logger) tag(tag string, color string) {
	if !emerald.ColorEnabled {
		color = ""
	}
	emerald.Print(color, "[", tag, "] ", emerald.Reset)
}

func (l Logger) directive() {
	emerald.Print("[")
	if emerald.ColorEnabled {
		emerald.Print(l.color)
	}
	emerald.Print(l.name, emerald.Reset, "] ")
}

// Log [<color><tag>] <msg>
func (l Logger) Log(a ...interface{}) {
	l.directive()
	emerald.Print(a...)
	emerald.Println(emerald.Reset)
}

// LogTag [<color><tag>] <grey><tag> <msg>
func (l Logger) LogTag(tag string, a ...interface{}) {
	l.directive()
	l.tag(tag, l.tagColor)
	emerald.Print(a...)
	emerald.Println(emerald.Reset)
}

func (l Logger) LogTagC(color string, tag string, a ...interface{}) {
	l.directive()
	l.tag(tag, color)
	emerald.Print(a...)
	emerald.Println(emerald.Reset)
}

// LogPath [<color><tag>] <grey><tag> <path1> <grey>-> <path2>
func (l Logger) LogPath(tag string, path1 string, path2 string) {
	l.directive()
	l.tag(tag, l.tagColor)
	emerald.Print(path1, emerald.LightBlack, " -> ", emerald.Reset, path2, "\n")
}

var colorError = emerald.ColorCode("red+b")

func Log(colorTag string, tag string, color string, a ...interface{}) {
	emerald.Print("[")
	if emerald.ColorEnabled {
		emerald.Print(colorTag)
	}
	emerald.Print(tag, emerald.Reset, "] ", color)
	emerald.Println(a...)
	emerald.Print(emerald.Reset)
}

func Error(a ...interface{}) {
	Log(colorError, "ERROR", emerald.Red, a...)
}
