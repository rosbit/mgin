package mgin

import (
	"github.com/urfave/negroni"
	"github.com/go-zoo/bone"
	"log"
	"os"
	"fmt"
	"net/http"
)

type MiniGin struct {
	RouterGroup
	n *negroni.Negroni
}

// type negroni.Handler interface {
//	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
// }
type Handler = negroni.Handler
// type HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
type HandlerFunc = negroni.HandlerFunc

func NewMgin(handlers ...Handler) *MiniGin {
	n := negroni.New()
	n.Use(negroni.NewRecovery())

	hasLogger := false
	for _, handler := range handlers {
		if _, ok := handler.(*negroni.Logger); ok {
			hasLogger = true
			n.Use(handler)
		}
	}
	if !hasLogger {
		n.Use(WithLogger("mgin"))
	}
	for _, handler := range handlers {
		if _, ok := handler.(*negroni.Logger); !ok {
			n.Use(handler)
		}
	}

	m := bone.New()
	n.UseHandler(m)
	return &MiniGin{
		RouterGroup: RouterGroup{
			basePath: "/",
			m: m,
		},
		n: n,
	}
}

func WithLogger(name string) Handler {
	logger := &negroni.Logger{ALogger: log.New(os.Stdout, fmt.Sprintf("[%s] ", name), 0)}
	logger.SetDateFormat(negroni.LoggerDefaultDateFormat)
	logger.SetFormat("{{.StartTime}} | {{.Status}} | \t {{.Duration}} | {{.Hostname}} | {{.Method}} {{.Request.RequestURI}}")
	return logger
}

func (h *MiniGin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.n.ServeHTTP(w, r)
}

func (h *MiniGin) Run(addr ...string) {
	h.n.Run(addr...)
}

func (h *MiniGin) WrapMiddleFunc(handlerFunc HandlerFunc) Handler {
	return HandlerFunc(handlerFunc)
}

func (h *MiniGin) Wrap(handler http.Handler) Handler {
	return negroni.Wrap(handler)
}

func (h *MiniGin) WrapFunc(handlerFunc http.HandlerFunc) Handler {
	return negroni.WrapFunc(handlerFunc)
}

func (hr *MiniGin) NotFoundHandler(h http.Handler) {
	hr.m.NotFound(h)
}
