package server

import (
	"github.com/chenyunchen/CrazyGo/error"
	"github.com/chenyunchen/CrazyGo/handlers"
	"html/template"
	"net/http"
	"path"
)

// Router
type server struct {
	router      map[string]map[string][]interface{}
	middlewares []func(handlers.Handler) handlers.Handler
}

func New() *server {
	return &server{make(map[string]map[string][]interface{}), make([]func(handlers.Handler) handlers.Handler, 0)}
}

func handleVerbs(method string, s *server, path string, fn func(http.ResponseWriter, *http.Request), middlewares []func(handlers.Handler) handlers.Handler) {
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

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for key, value := range s.router {
		if r.URL.Path == key {
			for k, v := range value {
				if r.Method == k {
					length := len(v)
					fn := v[0].(func(http.ResponseWriter, *http.Request))
					if len(s.middlewares) > 0 {
						if length > 1 {
							middlewares := make([]func(handlers.Handler) handlers.Handler, length-1)
							for i, h := range v[1:] {
								middlewares[i] = h.(func(handlers.Handler) handlers.Handler)
							}
							newM := append(s.middlewares, middlewares...)
							handlers.Merge(w, r, handlers.Handler{fn}, newM)
							return
						} else {
							handlers.Merge(w, r, handlers.Handler{fn}, s.middlewares)
							return
						}
					} else {
						if length > 1 {
							middlewares := make([]func(handlers.Handler) handlers.Handler, length-1)
							for i, h := range v[1:] {
								middlewares[i] = h.(func(handlers.Handler) handlers.Handler)
							}
							handlers.Merge(w, r, handlers.Handler{fn}, middlewares)
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

func (s *server) Use(fn func(handlers.Handler) handlers.Handler) {
	s.middlewares = append(s.middlewares, fn)
}

func (s *server) Get(path string, fn func(http.ResponseWriter, *http.Request), middlewares ...func(handlers.Handler) handlers.Handler) {
	handleVerbs("GET", s, path, fn, middlewares)
}

func (s *server) Post(path string, fn func(http.ResponseWriter, *http.Request), middlewares ...func(handlers.Handler) handlers.Handler) {
	handleVerbs("POST", s, path, fn, middlewares)
}

func (s *server) Put(path string, fn func(http.ResponseWriter, *http.Request), middlewares ...func(handlers.Handler) handlers.Handler) {
	handleVerbs("PUT", s, path, fn, middlewares)
}

func (s *server) Delete(path string, fn func(http.ResponseWriter, *http.Request), middlewares ...func(handlers.Handler) handlers.Handler) {
	handleVerbs("DELETE", s, path, fn, middlewares)
}

//HTML & JSON
func Render(w http.ResponseWriter, filename string, data interface{}) {
	filepath := path.Join("templates", filename)
	temp, err := template.ParseFiles(filepath)
	if err != nil {
		error.ResError(w, error.InternalServerError)
		return
	}
	err = temp.Execute(w, data)
	if err != nil {
		error.ResError(w, error.InternalServerError)
	}
}
