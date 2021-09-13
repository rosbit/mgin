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
// type negroni..HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
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
		n.Use(WithLogger("http-helper"))
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

func (h *MiniGin) Use(handler Handler) {
	h.n.Use(handler)
}

func (h *MiniGin) UseFunc(handlerFunc HandlerFunc) {
	h.n.UseFunc(handlerFunc)
}

func (h *MiniGin) UseHandler(handler http.Handler) {
	h.n.Use(negroni.Wrap(handler))
}

func (h *MiniGin) UseHandlerFunc(handlerFunc http.HandlerFunc) {
	h.n.UseHandlerFunc(handlerFunc)
}

func (hr *MiniGin) NotFoundHandler(h http.Handler) {
	hr.m.NotFound(h)
}
