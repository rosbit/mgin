package mgin

import (
	"io"
)

func bodyDumper(body io.Reader, dumper io.Writer) (reader io.Reader, deferFunc func()) {
	if dumper == nil {
		return body, func(){}
	}

	io.WriteString(dumper, "--- mgin dumping body begin ---\n")
	r := io.TeeReader(body, dumper)
	return r, func() {
		io.WriteString(dumper, "\n--- mgin dumping body end ---\n")
	}
}
