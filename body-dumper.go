package mgin

import (
	logr "github.com/rosbit/reader-logger"
	"io"
)

func bodyDumper(body io.Reader, dumper io.Writer) (reader io.Reader, deferFunc func()) {
	return logr.ReaderLogger(body, dumper, "mgin dumping body")
}
