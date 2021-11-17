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

func (l *Logger) Print(a ...interface{}) *Logger {
	emerald.Print(a...)
	emerald.Print(emerald.Reset)
	return l
}

func (l *Logger) Printf(format string, a ...interface{}) *Logger {
	emerald.Printf(format, a...)
	emerald.Print(emerald.Reset)
	return l
}

func (l *Logger) Println(a ...interface{}) *Logger {
	emerald.Println(a...)
	emerald.Print(emerald.Reset)
	return l
}

func (l *Logger) Tag(tag string) *Logger {
	l.tag(tag, l.tagColor)
	return l
}

func (l *Logger) TagC(color, tag string) *Logger {
	l.tag(tag, color)
	return l
}

func (l *Logger) Path(path1, path2 string) *Logger {
	emerald.Print(path1, emerald.LightBlack, " -> ", emerald.Reset, path2, "\n")
	return l
}

func (l *Logger) Log() *Logger {
	l.directive()
	return l
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

func (l Logger) LogPath(tag string, path1 string, path2 string) {
	l.directive()
	l.tag(tag, l.tagColor)
	emerald.Print(path1, emerald.LightBlack, " -> ", emerald.Reset, path2, "\n")
}

func (l Logger) LogPathC(color string, tag string, path1 string, path2 string) {
	l.directive()
	l.tag(tag, color)
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
