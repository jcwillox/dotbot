package log

import (
	"bytes"
	"github.com/jcwillox/emerald"
	"golang.org/x/term"
	"os"
)

type MaxLineWriter struct {
	buf      bytes.Buffer
	lines    *[]string
	maxLines int
	maxWidth int
}

func (w MaxLineWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		w.buf.WriteByte(b)
		if b == '\n' || w.buf.Len() == w.maxWidth {
			if b != '\n' {
				w.buf.WriteByte('\n')
			}
			line := w.buf.String()
			w.buf.Reset()
			if len(*w.lines) < w.maxLines {
				*w.lines = append(*w.lines, line)
				emerald.Print(line)
			} else {
				*w.lines = append((*w.lines)[1:], line)
				emerald.CursorUp(w.maxLines)
				for _, line := range *w.lines {
					emerald.Print("\x1b[2K", line)
				}
			}
		}
	}
	return len(p), nil
}

func NewMaxLineWriter(maxLines int) MaxLineWriter {
	arr := make([]string, 0, maxLines)
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 60
	}
	return MaxLineWriter{
		lines:    &arr,
		maxLines: maxLines,
		maxWidth: width,
	}
}
