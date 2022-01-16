package log

import (
	"fmt"
	"github.com/jcwillox/emerald"
	"github.com/mattn/go-colorable"
	"os"
)

var (
	Stderr = colorable.NewColorableStdout()
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

func (l *Logger) tag(tag string, color string) *Logger {
	if !emerald.ColorEnabled {
		color = ""
	}
	emerald.Print(color, "[", tag, "] ", emerald.Reset)
	return l
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
	return l.tag(tag, l.tagColor)
}

func (l *Logger) TagC(color, tag string) *Logger {
	return l.tag(tag, color)
}

func (l *Logger) Sudo(sudo ...bool) *Logger {
	return l.SudoC(emerald.Magenta, sudo...)
}

func (l *Logger) SudoC(color string, sudo ...bool) *Logger {
	if len(sudo) > 0 && sudo[0] {
		return l.tag("sudo", color)
	} else if _, present := os.LookupEnv("DOTBOT_SUDO"); present {
		return l.tag("sudo", color)
	}
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

// Warn  - program will always continue
// Error - recoverable, user can set these to be ignored
// Fatal - will always exit the program
// Panic - will always crash the program

func tag(color string, tag string) {
	fmt.Fprint(Stderr, "[", color, tag, emerald.Reset, "] ", color)
}

var (
	errorTag = func() {
		tag(emerald.LightRed, "ERROR")
	}
	warnTag = func() {
		tag(emerald.Yellow, "WARN")
	}
	fatalTag = func() {
		tag(emerald.Bold+emerald.LightRed, "FATAL")
	}
	panicTag = func() {
		tag(emerald.Bold+emerald.LightMagenta, "PANIC")
	}
)

func Warn(a ...interface{}) {
	warnTag()
	fmt.Fprint(Stderr, a...)
	fmt.Fprint(Stderr, emerald.Reset)
}

func Warnf(format string, a ...interface{}) {
	warnTag()
	fmt.Fprintf(Stderr, format, a...)
	fmt.Fprint(Stderr, emerald.Reset)
}

func Warnln(a ...interface{}) {
	Warn(a...)
	fmt.Fprintln(Stderr)
}

func Error(a ...interface{}) {
	errorTag()
	fmt.Fprint(Stderr, a...)
	fmt.Fprint(Stderr, emerald.Reset)
}

func Errorf(format string, a ...interface{}) {
	errorTag()
	fmt.Fprintf(Stderr, format, a...)
	fmt.Fprint(Stderr, emerald.Reset)
}

func Errorln(a ...interface{}) {
	Error(a...)
	fmt.Fprintln(Stderr)
}

func Fatal(a ...interface{}) {
	fatalTag()
	fmt.Fprint(Stderr, a...)
	fmt.Fprint(Stderr, emerald.Reset)
	os.Exit(1)
}

func Fatalf(format string, a ...interface{}) {
	fatalTag()
	fmt.Fprintf(Stderr, format, a...)
	fmt.Fprint(Stderr, emerald.Reset)
	os.Exit(1)
}

func Fatalln(a ...interface{}) {
	fatalTag()
	fmt.Fprint(Stderr, a...)
	fmt.Fprintln(Stderr, emerald.Reset)
	os.Exit(1)
}

func Panicln(a ...interface{}) {
	panicTag()
	s := fmt.Sprintln(a...)
	fmt.Fprint(Stderr, s, emerald.Reset)
	panic(s)
}
