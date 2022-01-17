package log

import (
	"bytes"
	"github.com/jcwillox/emerald"
)

type MaxLineWriter struct {
	lines *[]string
	buf   bytes.Buffer
}

const maxLines = 10

func (w MaxLineWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		w.buf.WriteByte(b)
		if b == '\n' {
			line := w.buf.String()
			w.buf.Reset()
			if len(*w.lines) < maxLines {
				*w.lines = append(*w.lines, line)
				emerald.Print(line)
			} else {
				*w.lines = append((*w.lines)[1:], line)
				emerald.CursorUp(maxLines)
				for _, line := range *w.lines {
					emerald.Print("\x1b[2K", line)
				}
			}
		}
	}
	return len(p), nil
}

func NewMaxLineWriter() MaxLineWriter {
	arr := make([]string, 0, maxLines)
	return MaxLineWriter{
		lines: &arr,
	}
}
