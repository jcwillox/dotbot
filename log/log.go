package log

import (
	"fmt"
	"github.com/jcwillox/emerald"
	"github.com/k0kubun/pp/v3"
	"github.com/mattn/go-colorable"
	"golang.org/x/term"
	"os"
	"strings"
)

var (
	EnableDebug = false
	Stderr      = colorable.NewColorableStdout()
	ColorSudo   = emerald.LightMagenta
	ColorNew    = emerald.LightYellow
	ColorDone   = emerald.LightBlack
)

var (
	ppLog = pp.New()
)

func init() {
	ppLog.SetOutput(Stderr)
	if debug := os.Getenv("DOTBOT_DEBUG"); debug == "true" || debug == "1" {
		EnableDebug = true
	}
}

func NewBasicLogger(name string) *Logger {
	return &Logger{
		name: name,
	}
}

type Logger struct {
	name     string
	ShowTime bool
}

func (l *Logger) tag(tag string, color string) *Logger {
	if !emerald.ColorEnabled {
		color = ""
	}
	emerald.Print(color, "[", tag, "] ", emerald.Reset)
	return l
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
	return l.tag(tag, ColorNew)
}

func (l *Logger) TagDone(tag string) *Logger {
	return l.tag(tag, ColorDone)
}

func (l *Logger) TagSudo(tag string, sudo ...bool) *Logger {
	if len(sudo) > 0 && sudo[0] {
		return l.tag(tag, ColorSudo)
	} else if _, present := os.LookupEnv("DOTBOT_SUDO"); present {
		return l.tag(tag, ColorSudo)
	}
	return l.tag(tag, ColorNew)
}

func (l *Logger) TagC(color, tag string) *Logger {
	return l.tag(tag, color)
}

func (l *Logger) Path(path1, path2 string) *Logger {
	emerald.Print(path1, emerald.LightBlack, " -> ", emerald.Reset, path2, "\n")
	return l
}

func Rule(msg string) {
	if !emerald.ColorEnabled {
		fmt.Println("──", msg, "──")
	} else {
		bar := "──"
		extra := ""
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			freeSpace := width - len(msg) - 2
			barLength := freeSpace / 2
			if barLength > 0 {
				extraLength := freeSpace % 2
				bar = strings.Repeat("─", barLength)
				extra = strings.Repeat("─", extraLength)
			}
		}
		emerald.Print(
			emerald.LightBlack, bar, emerald.Bold, emerald.Blue, " ", msg, " ",
			emerald.Reset, emerald.LightBlack, bar, extra, emerald.Reset, "\n",
		)
	}
}

// Warn  - program will always continue
// Error - recoverable, user can set these to be ignored
// Fatal - will always exit the program
// Panic - will always crash the program

func tag(color string, tag string) {
	fmt.Fprint(Stderr, "[", color, tag, emerald.Reset, "] ", color)
}

var (
	debugTag = func() {
		tag(emerald.LightCyan, "DEBUG")
	}
	warnTag = func() {
		tag(emerald.Yellow, "WARN")
	}
	errorTag = func() {
		tag(emerald.LightRed, "ERROR")
	}
	fatalTag = func() {
		tag(emerald.Bold+emerald.LightRed, "FATAL")
	}
	panicTag = func() {
		tag(emerald.Bold+emerald.LightMagenta, "PANIC")
	}
)

func Debug(a ...interface{}) {
	if !EnableDebug {
		return
	}
	debugTag()
	fmt.Fprint(Stderr, a...)
	fmt.Fprint(Stderr, emerald.Reset)
}

func Debugf(format string, a ...interface{}) {
	if !EnableDebug {
		return
	}
	debugTag()
	fmt.Fprintf(Stderr, format, a...)
	fmt.Fprint(Stderr, emerald.Reset)
}

func Debugln(a ...interface{}) {
	if !EnableDebug {
		return
	}
	debugTag()
	fmt.Fprintln(Stderr, a...)
	fmt.Fprint(Stderr, emerald.Reset)
}

func DebugPpln(a ...interface{}) {
	if !EnableDebug {
		return
	}
	debugTag()
	fmt.Fprint(Stderr, emerald.Reset)
	ppLog.Println(a)
}

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
	warnTag()
	fmt.Fprintln(Stderr, a...)
	fmt.Fprint(Stderr, emerald.Reset)
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
	errorTag()
	fmt.Fprintln(Stderr, a...)
	fmt.Fprint(Stderr, emerald.Reset)
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
	fmt.Fprintln(Stderr, a...)
	fmt.Fprint(Stderr, emerald.Reset)
	os.Exit(1)
}

func Panicln(a ...interface{}) {
	panicTag()
	s := fmt.Sprintln(a...)
	fmt.Fprint(Stderr, s, emerald.Reset)
	panic(s)
}
