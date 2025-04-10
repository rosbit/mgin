package mgin

import (
	"net/http"
)

func CreateCoresFreeHandler() Handler {
	return WrapMiddleFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Headers", "*")
		h.Set("Access-Control-Allow-Method", "POST, GET, PUT, OPTIONS, DELETE")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	})
}
