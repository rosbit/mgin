package mgin

import (
	logr "github.com/rosbit/reader-logger"
	"io"
	"os"
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

func CreateBodyDumpingHandler(dumper io.Writer, prompt ...string) Handler {
	return WrapMiddleFunc(logr.CreateBodyDumpingHandlerFunc(dumper, prompt...))
}
