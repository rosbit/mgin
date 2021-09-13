# mgin

A http framework to support router and argument processing

## Sample #1: handler with http.HandlerFunc

```go
package main

import (
    "github.com/rosbit/mgin"
    "net/http"
    "fmt"
)

func main() {
    r := mgin.NewMgin()

    r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
        c := mgin.NewHttpContext(w, r)
        c.String(http.StatusOK, "hello")
    })

    r.Get("/json/:msg", func(w http.ResponseWriter, r *http.Request) {
        c := mgin.NewHttpContext(w, r)
        msg := c.Param("msg")
        c.JSON(http.StatusOK, map[string]interface{} {
            "code": http.StatusOK,
            "msg": msg,
        })
    })

    r.Post("/json", func(w http.ResponseWriter, r *http.Request) {
        c := mgin.NewHttpContext(w, r)
        var i interface{}
        code, err := c.ReadJSON(&i)
        if err != nil {
            c.Error(code, err.Error())
            return
        }
        c.JSONPretty(http.StatusOK, i, " ")
    })

    r.Get("/jump", func(w http.ResponseWriter, r *http.Request) {
        c := mgin.NewHttpContext(w, r)
        url := c.QueryParam("u")
        if url == "" {
            c.Error(http.StatusBadRequest, "argument u expected")
            return
        }
        c.Redirect(http.StatusFound, url)
    })

    r.Post("/form/:name", func(w http.ResponseWriter, r *http.Request) {
        c := mgin.NewHttpContext(w, r)
        n := c.Param("name")
        v := c.FormValue(n)
        c.String(http.StatusOK, fmt.Sprintf("value of %s: %s\n", n, v))
    })

    r.Run()
    // or r.Run(":8080")
    // or http.ListenAndServe(":8080", r)
}
```

## Sample #2: handler with argument mgin.Context

```go
package main

import (
    "github.com/rosbit/http-mgin"
    "net/http"
    "fmt"
)

func main() {
    r := mgin.NewMgin()

    r.GET("/hello", func(c *mgin.Context) {
        c.String(http.StatusOK, "hello")
    })

    r.GET("/json/:msg", func(c *mgin.Context) {
        msg := c.Param("msg")
        c.JSON(http.StatusOK, map[string]interface{} {
            "code": http.StatusOK,
            "msg": msg,
        })
    })

    r.POST("/json", func(c *mgin.Context) {
        var i interface{}
        code, err := c.ReadJSON(&i)
        if err != nil {
            c.Error(code, err.Error())
            return
        }
        c.JSONPretty(http.StatusOK, i, " ")
    })

    r.GET("/jump", func(c *mgin.Context) {
        url := c.QueryParam("u")
        if url == "" {
            c.Error(http.StatusBadRequest, "argument u expected")
            return
        }
        c.Redirect(http.StatusFound, url)
    })

    r.POST("/form/:name", func(c *mgin.Context) {
        n := c.Param("name")
        v := c.FormValue(n)
        c.String(http.StatusOK, fmt.Sprintf("value of %s: %s\n", n, v))
    })

    r.Run()
    // or r.Run(":8080")
    // or http.ListenAndServe(":8080", r)
}
```
