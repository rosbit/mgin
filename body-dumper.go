package mgin

import (
	logr "github.com/rosbit/reader-logger"
	"io"
	"os"
)

func bodyDumper(body io.Reader, dumper io.Writer) (reader io.Reader, deferFunc func()) {
	var prompt string
	if dumper != nil {
		if f, ok := dumper.(*os.File); ok {
			if f == os.Stderr || f == os.Stdout {
				prompt = "mgin dumping body"
			}
		}
	}
	return logr.ReaderLogger(body, dumper, prompt)
}
