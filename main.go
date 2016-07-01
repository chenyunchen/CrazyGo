package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type test struct{}

func (h test) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test")
}

//func chain(ch ...http.Handler) http.Handler {
//	for c := range ch {
//
//	}
//}

func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func errorHandler(next http.Handler) http.Handler {
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

func main() {
	http.Handle("/", errorHandler(loggingHandler(test{})))
	http.ListenAndServe(":8080", nil)
}
