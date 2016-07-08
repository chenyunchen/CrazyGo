package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	next func(http.ResponseWriter, *http.Request)
}

// Router
// ==========================================================

type Server struct {
	router      map[string]map[string]func(http.ResponseWriter, *http.Request)
	middlewares []func(Handler) Handler
}

func New() *Server {
	return &Server{make(map[string]map[string]func(http.ResponseWriter, *http.Request)), make([]func(Handler) Handler, 0)}
}

func handleVerbs(method string, s *Server, path string, fn func(http.ResponseWriter, *http.Request)) {
	_, ok := s.router[path]
	if !ok {
		s.router[path] = make(map[string]func(http.ResponseWriter, *http.Request))
	}
	s.router[path][method] = fn
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for key, value := range s.router {
		if r.URL.Path == key {
			for k, v := range value {
				if r.Method == k {
					if len(s.middlewares) > 0 {
						Merge(w, r, Handler{v}, s.middlewares)
						return
					} else {
						v(w, r)
						return
					}
				}
			}
		}
	}
	http.NotFound(w, r)
	return
}

func (s *Server) Use(fn func(Handler) Handler) {
	s.middlewares = append(s.middlewares, fn)
}

func (s *Server) Get(path string, fn func(http.ResponseWriter, *http.Request)) {
	handleVerbs("GET", s, path, fn)
}

func (s *Server) Post(path string, fn func(http.ResponseWriter, *http.Request)) {
	handleVerbs("POST", s, path, fn)
}

func (s *Server) Put(path string, fn func(http.ResponseWriter, *http.Request)) {
	handleVerbs("PUT", s, path, fn)
}

func (s *Server) Delete(path string, fn func(http.ResponseWriter, *http.Request)) {
	handleVerbs("DELETE", s, path, fn)
}

// ==========================================================

// Context
// ----------------------------------------------------------
type context struct {
	data map[*http.Request]map[interface{}]interface{}
}

var c context = context{make(map[*http.Request]map[interface{}]interface{})}

func (con *context) Set(r *http.Request, key, value interface{}) {
	_, ok := con.data[r]
	if !ok {
		con.data[r] = make(map[interface{}]interface{})
	}
	con.data[r][key] = value

}
func (con context) Get(r *http.Request, key interface{}) interface{} {
	return con.data[r][key]
}

// ----------------------------------------------------------

// Merge Multiple Handler
// ==========================================================
func Merge(w http.ResponseWriter, r *http.Request, h Handler, middlewares []func(Handler) Handler) {
	var chain Handler
	chain = middlewares[0](h)
	size := len(middlewares)
	for _, middleware := range middlewares[size-1:] {
		chain = middleware(chain)
	}
	chain.next(w, r)
}

// ==========================================================

// Define Handler
// ----------------------------------------------------------
func logH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		h.next(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return Handler{fn}
}

func errorH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Error] %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()

		h.next(w, r)
	}
	return Handler{fn}
}

//Auth Example
func authH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Do something check user identity
		user := make(map[string]string)
		user["name"] = "test"
		c.Set(r, "user", user)
		h.next(w, r)
	}
	return Handler{fn}
}

// ----------------------------------------------------------

// Test Example for Handlers
func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test")
	user := c.Get(r, "user").(map[string]string)
	fmt.Println("Name:", user["name"])
}

func main() {
	app := New()
	app.Use(logH)
	app.Use(errorH)
	app.Use(authH)
	app.Get("/test", test)
	http.ListenAndServe(":8080", app)
}
