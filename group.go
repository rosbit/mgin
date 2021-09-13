package mgin

import (
	"github.com/go-zoo/bone"
	"path"
	"net/http"
)

type RouterGroup struct {
	basePath string
	m *bone.Mux
}

func (g *RouterGroup) Group(relPath string) *RouterGroup {
	return &RouterGroup{
		basePath: g.toAbsPath(relPath),
		m: g.m,
	}
}

func (g *RouterGroup) toAbsPath(relPath string) string {
	return toAbsPath(g.basePath, relPath)
}

func toAbsPath(absPath, relPath string) string {
	if len(relPath) == 0 {
		return absPath
	}

	resPath := path.Join(absPath, relPath)
	if lastChar(relPath) == '/' && lastChar(resPath) != '/' {
		return resPath + "/"
	}
	return resPath
}

func lastChar(s string) byte {
	l := len(s)
	if l == 0 {
		panic("null string not allowed")
	}
	return s[l-1]
}

func (g *RouterGroup) Get(pattern string, h http.HandlerFunc) {
	g.m.Get(g.toAbsPath(pattern), h)
}

// ----- HandlerFunc ----
func (g *RouterGroup) Put(path string, h http.HandlerFunc) {
	g.m.Put(g.toAbsPath(path), h)
}

func (g *RouterGroup) Post(path string, h http.HandlerFunc) {
	g.m.Post(g.toAbsPath(path), h)
}

func (g *RouterGroup) Patch(path string, h http.HandlerFunc) {
	g.m.Patch(g.toAbsPath(path), h)
}

func (g *RouterGroup) Head(path string, h http.HandlerFunc) {
	g.m.Head(g.toAbsPath(path), h)
}

func (g *RouterGroup) Options(path string, h http.HandlerFunc) {
	g.m.Options(g.toAbsPath(path), h)
}

func (g *RouterGroup) Delete(path string, h http.HandlerFunc) {
	g.m.Delete(g.toAbsPath(path), h)
}

// ------ context -----
func (g *RouterGroup) GET(pattern string, h func(c *Context)) {
	g.m.Get(g.toAbsPath(pattern), unwrap(h))
}

func (g *RouterGroup) PUT(path string, h func(c *Context)) {
	g.m.Put(g.toAbsPath(path), unwrap(h))
}

func (g *RouterGroup) POST(path string, h func(c *Context)) {
	g.m.Post(g.toAbsPath(path), unwrap(h))
}

func (g *RouterGroup) PATCH(path string, h func(c *Context)) {
	g.m.Patch(g.toAbsPath(path), unwrap(h))
}

func (g *RouterGroup) OPTIONS(path string, h func(c *Context)) {
	g.m.Options(g.toAbsPath(path), unwrap(h))
}

func (g *RouterGroup) HEAD(path string, h func(c *Context)) {
	g.m.Head(g.toAbsPath(path), unwrap(h))
}

func (g *RouterGroup) DELETE(path string, h func(c *Context)) {
	g.m.Delete(g.toAbsPath(path), unwrap(h))
}
