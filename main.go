package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
)

// Error
// ----------------------------------------------------------
type Errors struct {
	Errors []*Error `json:"errors"`
}
type Error struct {
	Type   string `json:"type"`
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

var (
	InternalServerError = &Error{"internal_server_error", 500, "Internal Server Error", "OMG"}
	NotAcceptableError  = &Error{"not_acceptable_error", 406, "Not Acceptable", "OMG"}
	CrazyError          = &Error{"crazy_error", 000, "I'm Crazy.", "◓ Д ◒"}
)

func ResError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}

// ----------------------------------------------------------

//HTML & JSON
// ----------------------------------------------------------
func Render(w http.ResponseWriter, filename string, data interface{}) {
	filepath := path.Join("templates", filename)
	temp, err := template.ParseFiles(filepath)
	if err != nil {
		ResError(w, InternalServerError)
		return
	}
	err = temp.Execute(w, data)
	if err != nil {
		ResError(w, InternalServerError)
	}
}

// ----------------------------------------------------------

// Router
// ==========================================================
type Server struct {
	router      map[string]map[string][]interface{}
	middlewares []func(Handler) Handler
}

func New() *Server {
	return &Server{make(map[string]map[string][]interface{}), make([]func(Handler) Handler, 0)}
}

func handleVerbs(method string, s *Server, path string, fn func(http.ResponseWriter, *http.Request), middlewares []func(Handler) Handler) {
	_, ok := s.router[path]
	if !ok {
		s.router[path] = make(map[string][]interface{})
	}
	_, ok = s.router[path][method]
	if !ok {
		s.router[path][method] = make([]interface{}, 0)
	}
	s.router[path][method] = append(s.router[path][method], fn)
	for _, middleware := range middlewares {
		s.router[path][method] = append(s.router[path][method], middleware)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for key, value := range s.router {
		if r.URL.Path == key {
			for k, v := range value {
				if r.Method == k {
					length := len(v)
					fn := v[0].(func(http.ResponseWriter, *http.Request))
					if len(s.middlewares) > 0 {
						if length > 1 {
							middlewares := make([]func(Handler) Handler, length-1)
							for i, h := range v[1:] {
								middlewares[i] = h.(func(Handler) Handler)
							}
							newM := append(s.middlewares, middlewares...)
							Merge(w, r, Handler{fn}, newM)
							return
						} else {
							Merge(w, r, Handler{fn}, s.middlewares)
							return
						}
					} else {
						if length > 1 {
							middlewares := make([]func(Handler) Handler, length-1)
							for i, h := range v[1:] {
								middlewares[i] = h.(func(Handler) Handler)
							}
							Merge(w, r, Handler{fn}, middlewares)
							return
						} else {
							fn(w, r)
							return
						}
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

func (s *Server) Get(path string, fn func(http.ResponseWriter, *http.Request), middlewares ...func(Handler) Handler) {
	handleVerbs("GET", s, path, fn, middlewares)
}

func (s *Server) Post(path string, fn func(http.ResponseWriter, *http.Request), middlewares ...func(Handler) Handler) {
	handleVerbs("POST", s, path, fn, middlewares)
}

func (s *Server) Put(path string, fn func(http.ResponseWriter, *http.Request), middlewares ...func(Handler) Handler) {
	handleVerbs("PUT", s, path, fn, middlewares)
}

func (s *Server) Delete(path string, fn func(http.ResponseWriter, *http.Request), middlewares ...func(Handler) Handler) {
	handleVerbs("DELETE", s, path, fn, middlewares)
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
type Handler struct {
	next func(http.ResponseWriter, *http.Request)
}

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
				ResError(w, InternalServerError)
			}
		}()

		h.next(w, r)
	}
	return Handler{fn}
}

func checkHeaderH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.api+json" {
			ResError(w, NotAcceptableError)
			return
		}
		h.next(w, r)
	}
	return Handler{fn}
}

func bodyParseH(h Handler) Handler {
	// curl -H 'Accept: application/vnd.api+json' -d '{"id":1234567, "first_name":"◓ Д ◒", "last_name":"ˊ● ω ●ˋ", "middle_name":""}' http://localhost:8080/postjson
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := User{}
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			ResError(w, NotAcceptableError)
			return
		}
		c.Set(r, "body", user)
		h.next(w, r)
	}
	return Handler{fn}
}

//Auth Example
type User struct {
	Id         int    `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name,omitempty"`
}

func authH(h Handler) Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Do something check user identity
		user := User{1234567, "◓ Д ◒", "ˊ● ω ●ˋ", ""}
		c.Set(r, "user", user)
		h.next(w, r)
	}
	return Handler{fn}
}

// ----------------------------------------------------------

// Test Example for Handlers
func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test")
}

// Test JSON Example for Handlers
func testJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	user := c.Get(r, "user")
	json.NewEncoder(w).Encode(user)
}

// Test Error Example for Handlers
func testError(w http.ResponseWriter, r *http.Request) {
	ResError(w, CrazyError)
}

// Test Error Example for Handlers
func postJSON(w http.ResponseWriter, r *http.Request) {
	user := c.Get(r, "body").(User)
	fmt.Println(user.Id)
	fmt.Println(user.FirstName)
	fmt.Println(user.LastName)
}

// Test HTML Example for Handlers
func getHTML(w http.ResponseWriter, r *http.Request) {
	user := c.Get(r, "user")
	Render(w, "index.html", user)
}

func main() {
	app := New()
	app.Use(logH)
	app.Use(errorH)
	app.Use(bodyParseH)
	app.Get("/test", test)
	app.Get("/jsontest", testJSON, authH)
	app.Get("/jsonerror", testError)
	app.Post("/postjson", postJSON)
	app.Get("/gethtml", getHTML, authH)
	http.ListenAndServe(":8080", app)
}
