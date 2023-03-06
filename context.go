package mgin

import (
	"github.com/go-playground/validator/v10"
	"github.com/gin-contrib/sse"
	"github.com/go-zoo/bone"
	"mime/multipart"
	"encoding/json"
	"path/filepath"
	"net/http"
	"net/url"
	"strings"
	"reflect"
	"strconv"
	"context"
	"io"
	"os"
	"fmt"
)

const (
	defaultMemory = 32 << 20 // 32 MB
	indexPage = "index.html"
	headerContentType = "Content-Type"
	headerContentDisposition = "Content-Disposition"
	mimeMultipartForm = "multipart/form-data"
	mimeTextPlainCharsetUTF8 = "text/plain;charset=UTF-8"
	mimeApplicationJSONCharsetUTF8 = "application/json;charset=UTF-8"
)

type Context struct {
	w  http.ResponseWriter
	r *http.Request
	p  map[string]string
	q  url.Values
}

type ContextHandlerFunc func(c *Context)

func unwrap(handlerFunc ContextHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := NewHttpContext(w, r)
		handlerFunc(c)
	}
}

func NewHttpContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{w:w, r:r}
}

func (c *Context) Request() *http.Request {
	return c.r
}

func (c *Context) Response() http.ResponseWriter {
	return c.w
}

func (c *Context) Param(name string) string {
	if c.p == nil {
		c.p = bone.GetAllValues(c.r)
	}
	if len(c.p) == 0 {
		return ""
	}
	if v, ok := c.p[name]; ok {
		return v
	}
	return ""
}

// read values from path, query string, form, header or cookie, store them in a struct specified by vals.
// param name is specified by field tag. e.g.
//
// var vals struct {
//    V1 int    `path:"v1" optional`       // read path param "v1" optionally, any integer type (u)int8/16/32/64 is acceptable
//    V2 bool   `query:"v2" ignore-error`  // read query param "v2=xxx", ignore error occurring
//    V3 string `form:"v3"`                // read form param "v3=xxx", type string can be replace with []byte
//    V4 int    `header:"Content-Length"`  // read header "Content-Length"
//    V5 []byte `cookie:"v5" optional`     // read cookie "v4" opiontally
// }
// if status, err := c.ReadParams(&vals); err != nil {
//    c.Error(status, err.Error())
//    return
// }
func (c *Context) ReadParams(vals interface{}) (status int, err error) {
	return c.readParams(vals, false)
}

// read values from path, query string, form, header or cookie, store them in a struct specified by vals,
// then values are validated by using "github.com/go-playground/validator/v10".
// param name is specified by field tag. e.g.
//
// var vals struct {
//    V1 int    `path:"v1" validate:"gt=0"`     // read path param "v1", the value must be greater than 0
//    V2 bool   `query:"v2"`                    // read query param "v2=xxx"
//    V3 string `form:"v3" validate:"required"` // read form param "v3=xxx",
//    V4 int    `header:"Content-Length"`       // read header "Content-Length"
//    V5 []byte `cookie:"v5"`                   // read cookie "v4"
// }
// if status, err := c.ReadAndValidate(&vals); err != nil {
//    c.Error(status, err.Error())
//    return
// }
func (c *Context) ReadAndValidate(vals interface{}) (status int, err error) {
	if status, err = c.readParams(vals, true); err != nil {
		return
	}
	v := validator.New()
	if err = v.Struct(vals); err != nil {
		status = http.StatusBadRequest
	}
	return
}

func (c *Context) readParams(vals interface{}, onlyRead bool) (status int, err error) {
	type tagHandler struct {
		tagName string
		getVal  func(string)string
	}

	if vals == nil {
		return http.StatusOK, nil
	}
	p := reflect.ValueOf(vals)
	if p.Kind() != reflect.Ptr {
		return http.StatusInternalServerError, fmt.Errorf("vals must be pointer")
	}
	v := p.Elem() // struct Value
	if v.Kind() != reflect.Struct {
		return http.StatusInternalServerError, fmt.Errorf("vals must be pointer of struct")
	}
	t := v.Type() // struct Type
	n := t.NumField()
	tagHandlers := []tagHandler{
		tagHandler{"path",  c.Param},
		tagHandler{"query", c.QueryParam},
		tagHandler{"form",  c.FormValue},
		tagHandler{"header",c.Header},
		tagHandler{"cookie",c.CookieValue},
	}
	for i:=0; i<n; i++ {
		field := t.Field(i) // StructField
		val := ""
		for _, tagHandler := range tagHandlers {
			if tag, ok := field.Tag.Lookup(tagHandler.tagName); ok {
				val = tagHandler.getVal(tag)
				break
			}
		}
		if !onlyRead {
			if len(val) == 0 {
				if _, optional := field.Tag.Lookup("optional"); optional {
					continue
				}
				return http.StatusBadRequest, fmt.Errorf("no value specified for field %s", field.Name)
			}
		}
		_, ignoreError := field.Tag.Lookup("ignore-error")
		ignoreError = !onlyRead && ignoreError

		fv := v.Field(i) // field Value
		ft := field.Type // field Type
		switch ft.Kind() {
		case reflect.String:
			fv.SetString(val)
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
			i, err := strconv.ParseInt(val, 10, ft.Bits())
			if err != nil {
				if !ignoreError { return http.StatusBadRequest, err }
				continue
			}
			fv.SetInt(i)
		case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
			i, err := strconv.ParseUint(val, 10, ft.Bits())
			if err != nil {
				if !ignoreError { return http.StatusBadRequest, err }
				continue
			}
			fv.SetUint(i)
		case reflect.Float64,reflect.Float32:
			f, err := strconv.ParseFloat(val, ft.Bits())
			if err != nil {
				if !ignoreError { return http.StatusBadRequest, err }
				continue
			}
			fv.SetFloat(f)
		case reflect.Bool:
			b, err := strconv.ParseBool(val)
			if err != nil {
				if !ignoreError { return http.StatusBadRequest, err }
				continue
			}
			fv.SetBool(b)
		case reflect.Slice:
			if ft.Elem().Kind() == reflect.Uint8 {
				fv.SetBytes([]byte(val))
				break
			}
			fallthrough
		default:
			return http.StatusNotImplemented, fmt.Errorf("value of type %s not implemented", ft.Name())
		}
	}
	return http.StatusOK, nil
}

func (c *Context) QueryParam(name string) string {
	if c.q == nil {
		c.q = c.r.URL.Query()
	}
	return c.q.Get(name)
}

func (c *Context) GetQueryArray(name string) ([]string, bool) {
	if c.q == nil {
		c.q = c.r.URL.Query()
	}
	vals, ok := c.q[name]
	return vals, ok
}

func (c *Context) GetQueryParam(name string) (string, bool) {
	if c.q == nil {
		c.q = c.r.URL.Query()
	}
	if vals, ok := c.q[name]; ok {
		return vals[0], true
	} else {
		return "", false
	}
}

func (c *Context) QueryParams() url.Values {
	if c.q == nil {
		c.q = c.r.URL.Query()
	}
	return c.q
}

func (c *Context) QueryString() string {
	return c.r.URL.RawQuery
}

func (c *Context) FormValue(name string) string {
	return c.r.FormValue(name)
}

func (c *Context) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.r.Header.Get(headerContentType), mimeMultipartForm) {
		if err := c.r.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.r.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.r.Form, nil
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.r.FormFile(name)
	return fh, err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.r.ParseMultipartForm(defaultMemory)
	return c.r.MultipartForm, err
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.r.Cookie(name)
}

func (c *Context) CookieValue(n string) string {
	if ck, err := c.Cookie(n); err != nil {
		return ""
	} else {
		return ck.Value
	}
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.w, cookie)
}

func (c *Context) Cookies() []*http.Cookie {
	return c.r.Cookies()
}

func (c *Context) Header(name string) string {
	return c.r.Header.Get(name)
}

func (c *Context) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

func (c *Context) AddHeader(key, value string) {
	c.w.Header().Add(key, value)
}

func (c *Context) writeContentType(contentType string) {
	c.w.Header().Set(headerContentType, contentType)
}

func (c *Context) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.w.WriteHeader(code)
	_, err = c.w.Write(b)
	return
}

func (c *Context) String(code int, s string) error {
	return c.Blob(code, mimeTextPlainCharsetUTF8, []byte(s))
}

func (c *Context) json(code int, i interface{}, indent string) error {
	enc := json.NewEncoder(c.w)
	enc.SetEscapeHTML(false)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	c.writeContentType(mimeApplicationJSONCharsetUTF8)
	c.w.WriteHeader(code)
	return enc.Encode(i)
}

func (c *Context) JSON(code int, i interface{}) error {
	return c.json(code, i, "")
}

func (c *Context) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, mimeApplicationJSONCharsetUTF8, b)
}

func (c *Context) JSONPretty(code int, i interface{}, indent string) (err error) {
	return c.json(code, i, indent)
}

func (c *Context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.w.WriteHeader(code)
	_, err = io.Copy(c.w, r)
	return
}

func NotFoundHandler(c *Context) (err error) {
	return c.Error(http.StatusNotFound, "File not found")
}

func (c *Context) Error(code int, msg string) (err error) {
	return c.JSON(code, map[string]interface{}{"code":code, "msg":msg})
}

func (c *Context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return NotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return NotFoundHandler(c)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(c.w, c.r, fi.Name(), fi.ModTime(), f)
	return
}

func (c *Context) contentDisposition(file, name, dispositionType string) error {
	c.w.Header().Set(headerContentDisposition, fmt.Sprintf("%s;filename=%s", dispositionType, name))
	return c.File(file)
}

func (c *Context) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *Context) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return fmt.Errorf("ErrInvalidRedirectCode %d", code)
	}
	http.Redirect(c.w, c.r, url, code)
	return nil
}

func (c *Context) ReadJSON(res interface{}, dumper ...io.Writer) (code int, err error) {
	if c.r.Body == nil {
		return http.StatusBadRequest, fmt.Errorf("bad request")
	}
	defer c.r.Body.Close()

	var realDumper io.Writer
	if len(dumper) > 0 && dumper[0] != nil {
		realDumper = dumper[0]
	}
	reqBody, deferFunc := bodyDumper(c.r.Body, realDumper)
	defer deferFunc()

	if err = json.NewDecoder(reqBody).Decode(res); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func (c *Context) ReadAndValidateJSON(res interface{}, dumper ...io.Writer) (code int, err error) {
	if code, err = c.ReadJSON(res, dumper...); err != nil {
		return
	}
	v := validator.New()
	if err = v.Struct(res); err != nil {
		code = http.StatusBadRequest
	}
	return
}

func (c *Context) SSEvent(name string, message interface{}) {
	sse.Event{Event: name, Data: message}.Render(c.w)
}

func (c *Context) Context() context.Context {
	return c.r.Context()
}

func (c *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(c.Header("Connection")), "upgrade") &&
		strings.EqualFold(c.Header("Upgrade"), "websocket") {
		return true
	}
	return false
}
