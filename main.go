package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Handler func(next http.Handler) http.Handler

// Merge Multiple Handler
// ==========================================================
type Execute struct {
	fn func(http.ResponseWriter, *http.Request)
}

func (h Execute) Check(x []Handler) http.Handler {
	var chain http.Handler
	chain = http.HandlerFunc(h.fn)
	for _, fn := range x {
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

// ----------------------------------------------------------

// Test Example for Handlers
func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test")
}

func main() {
	http.Handle("/", Execute{test}.Check([]Handler{logH, errorH}))
	http.ListenAndServe(":8080", nil)
}
