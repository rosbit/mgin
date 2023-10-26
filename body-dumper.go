package mgin

import (
	logr "github.com/rosbit/reader-logger"
	"io"
	"os"
	"net/http"
)

func bodyDumper(body io.Reader, dumper io.Writer, prompts ...string) (reader io.Reader, deferFunc func()) {
	var prompt string
	if len(prompts) > 0 {
		prompt = prompts[0]
	}

	if dumper != nil {
		if f, ok := dumper.(*os.File); ok {
			if f == os.Stderr || f == os.Stdout {
				if len(prompt) == 0 {
					prompt = "mgin dumping body"
				}
			}
		}
	}
	return logr.ReaderLogger(body, dumper, prompt)
}

type nopCloser struct {
	io.Reader
	deferFunc func()
}
func (rc *nopCloser) Close() error {
	rc.deferFunc()
	return nil
}
func wrapNopCloser(r io.Reader, deferFunc func()) *nopCloser {
	return &nopCloser{
		Reader: r,
		deferFunc: deferFunc,
	}
}

func CreateBodyDumpingHandler(dumper io.Writer, prompt ...string) Handler {
	return WrapMiddleFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if r.Body == nil {
			next(rw, r)
			return
		}
		nr, deferFunc := bodyDumper(r.Body, dumper, prompt...)
		r.Body = wrapNopCloser(nr, deferFunc)
		next(rw, r)
	})
}
