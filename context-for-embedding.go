// only used for embedding script.

package mgin

import (
	"encoding/json"
	"reflect"
	"unsafe"
	"io"
)

func (c *Context) ErrorString(err error) (errStr string) {
	if err != nil {
		errStr = err.Error()
	}
	return
}

func (c *Context) ReadBody() (body string, errStr string ) {
	if c.r.Body == nil {
		errStr = "bad request"
		return
	}
	defer c.r.Body.Close()
	b, e := io.ReadAll(c.r.Body)
	if e != nil {
		errStr = e.Error()
		return
	}
	bytes2str(&b, &body)
	return
}

func (c *Context) ReadJSONBody() (body interface{}, errStr string ) {
	if c.r.Body == nil {
		errStr = "bad request"
		return
	}
	defer c.r.Body.Close()
	if e := json.NewDecoder(c.r.Body).Decode(&body); e != nil {
		errStr = e.Error()
		return
	}
	return
}

func (c *Context) Json(code int, i interface{}) (errStr string) {
	if err := c.JSON(code, i); err != nil {
		errStr = err.Error()
	}
	return
}

func (c *Context) JsonBlob(code int, s string) (errStr string) {
	var b []byte
	str2bytes(&s, &b)
	if err := c.JSONBlob(code, b); err != nil {
		errStr = err.Error()
	}
	return
}

func (c *Context) StringBlob(code int, ct string, s string) (errStr string) {
	var b []byte
	str2bytes(&s, &b)
	if err := c.Blob(code, ct, b); err != nil {
		errStr = err.Error()
	}
	return
}

func (c *Context) WriteChunk(s string) (bytesWriten int, errStr string) {
	var b []byte
	str2bytes(&s, &b)
	var err error
	if bytesWriten, err = c.Write(b); err != nil {
		errStr = err.Error()
	}
	return
}

func bytes2str(b *[]byte, s *string) {
	bs := (*reflect.SliceHeader)(unsafe.Pointer(b))
	v := (*reflect.StringHeader)(unsafe.Pointer(s))
	v.Data = bs.Data
	v.Len  = bs.Len
}

func str2bytes(s *string, b *[]byte) {
	v := (*reflect.StringHeader)(unsafe.Pointer(s))
	bs := (*reflect.SliceHeader)(unsafe.Pointer(b))
	bs.Data = v.Data
	bs.Len  = v.Len
	bs.Cap  = v.Len
}
