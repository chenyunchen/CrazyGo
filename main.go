package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Handler func(next http.Handler) http.Handler

// Router
// ==========================================================

type Server struct {
	router map[string]map[string][]func(http.ResponseWriter, *http.Request)
}

func New() *Server {
	return &Server{make(map[string]map[string][]func(http.ResponseWriter, *http.Request))}
}

func handleVerbs(method string, s *Server, path string, fn func(http.ResponseWriter, *http.Request)) {
	_, ok := s.router[path]
	if !ok {
		s.router[path] = make(map[string][]func(http.ResponseWriter, *http.Request))
	}
	_, ok = s.router[path][method]
	if !ok {
		s.router[path][method] = make([]func(http.ResponseWriter, *http.Request), 0)
	}
	s.router[path][method] = append(s.router[path][method], fn)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for key, value := range s.router {
		if r.URL.Path == key {
			for k, v := range value {
				if r.Method == k {
					v[0](w, r)
					return
				}
			}
		}
	}
	http.NotFound(w, r)
	return
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
type Execute struct {
	fn func(http.ResponseWriter, *http.Request)
}

func (ex Execute) Check(hs ...Handler) http.Handler {
	var chain http.Handler
	chain = http.HandlerFunc(ex.fn)
	for _, fn := range hs {
		chain = fn(chain)
	}
	return chain
}

// ==========================================================

// Define Handler
// ----------------------------------------------------------
func logH(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func errorH(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Error] %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

//Auth Example
func authH(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Do something check user identity
		user := make(map[string]string)
		user["name"] = "test"
		c.Set(r, "user", user)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ----------------------------------------------------------

// Test Example for Handlers
func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test")
	//user := c.Get(r, "user").(map[string]string)
	//fmt.Println("Name:", user["name"])
}

func main() {
	app := New()
	app.Get("/test", test)
	//http.Handle("/test", Execute{test}.Check(logH, errorH, authH))
	http.ListenAndServe(":8080", app)
}
