package handlers

import (
	"encoding/json"
	"github.com/chenyunchen/CrazyGo/context"
	"github.com/chenyunchen/CrazyGo/error"
	"log"
	"net/http"
	"reflect"
	"time"
)

// Define Handler
type Handler struct {
	Next func(http.ResponseWriter, *http.Request)
}

// Define for lastwares' Handler
var Exit Handler = Handler{func(w http.ResponseWriter, r *http.Request) { return }}

// Merge Multiple Handler
func Merge(w http.ResponseWriter, r *http.Request, h Handler, middlewares []func(Handler) Handler) {
	var chain Handler
	chain = middlewares[0](h)
	size := len(middlewares)
	for _, middleware := range middlewares[size-1:] {
		chain = middleware(chain)
	}
	chain.Next(w, r)
}

func ClearH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		context.Clear(r)
		h.Next(w, r)
	}
	return Handler{fn}
}

func LogH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		h.Next(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return Handler{fn}
}

func ErrorH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Error] %+v", err)
				error.ResError(w, error.InternalServerError)
			}
		}()

		h.Next(w, r)
	}
	return Handler{fn}
}

func CheckHeaderH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.api+json" {
			error.ResError(w, error.NotAcceptableError)
			return
		}
		h.Next(w, r)
	}
	return Handler{fn}
}

func BodyParseH(obj interface{}) func(Handler) Handler {
	// curl -H 'Accept: application/vnd.api+json' -d '{"id":1234567, "first_name":"◓ Д ◒", "last_name":"ˊ● ω ●ˋ", "middle_name":""}' http://localhost:8080/postjson
	t := reflect.TypeOf(obj)
	outerFn := func(h Handler) Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			val := reflect.New(t).Interface()
			err := json.NewDecoder(r.Body).Decode(val)
			if err != nil {
				error.ResError(w, error.NotAcceptableError)
				return
			}
			context.Set(r, "body", val)
			h.Next(w, r)
		}
		return Handler{fn}
	}
	return outerFn
}
