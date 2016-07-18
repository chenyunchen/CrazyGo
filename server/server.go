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
	lastwares   map[string][]func(handlers.Handler) handlers.Handler
}

func New() *server {
	return &server{make(map[string]map[string][]interface{}), make([]func(handlers.Handler) handlers.Handler, 0), make(map[string][]func(handlers.Handler) handlers.Handler)}
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
	v := s.router[r.URL.Path][r.Method]
	length := len(v)
	if length == 0 {
		http.NotFound(w, r)
		return
	}
	fn := v[0].(func(http.ResponseWriter, *http.Request))
	if len(s.middlewares) > 0 {
		if length > 1 {
			middlewares := make([]func(handlers.Handler) handlers.Handler, length-1)
			for i, h := range v[1:] {
				middlewares[i] = h.(func(handlers.Handler) handlers.Handler)
			}
			newM := append(s.middlewares, middlewares...)
			handlers.Merge(w, r, handlers.Handler{fn}, newM)
			if len(s.lastwares[r.URL.Path]) != 0 {
				handlers.Merge(w, r, handlers.Exit, s.lastwares[r.URL.Path])
			}
			return
		} else {
			handlers.Merge(w, r, handlers.Handler{fn}, s.middlewares)
			if len(s.lastwares[r.URL.Path]) != 0 {
				handlers.Merge(w, r, handlers.Exit, s.lastwares[r.URL.Path])
			}
			return
		}
	} else {
		if length > 1 {
			middlewares := make([]func(handlers.Handler) handlers.Handler, length-1)
			for i, h := range v[1:] {
				middlewares[i] = h.(func(handlers.Handler) handlers.Handler)
			}
			handlers.Merge(w, r, handlers.Handler{fn}, middlewares)
			if len(s.lastwares[r.URL.Path]) != 0 {
				handlers.Merge(w, r, handlers.Exit, s.lastwares[r.URL.Path])
			}
			return
		} else {
			fn(w, r)
			if len(s.lastwares[r.URL.Path]) != 0 {
				handlers.Merge(w, r, handlers.Exit, s.lastwares[r.URL.Path])
			}
			return
		}
	}
}

func (s *server) Use(fn func(handlers.Handler) handlers.Handler) {
	if len(s.router) == 0 {
		s.middlewares = append(s.middlewares, fn)
	} else {
		for path, _ := range s.router {
			_, ok := s.lastwares[path]
			if !ok {
				s.lastwares[path] = make([]func(handlers.Handler) handlers.Handler, 0)
			}
			s.lastwares[path] = append(s.lastwares[path], fn)
		}
	}
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
