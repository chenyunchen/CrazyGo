package main

import (
	"encoding/json"
	"fmt"
	"github.com/chenyunchen/CrazyGo/context"
	"github.com/chenyunchen/CrazyGo/error"
	"github.com/chenyunchen/CrazyGo/handlers"
	"github.com/chenyunchen/CrazyGo/server"
	"net/http"
)

//Auth Example
type user struct {
	Id         int    `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name,omitempty"`
}

func authH(h handlers.Handler) handlers.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Do something check user identity
		u := user{1234567, "◓ Д ◒", "ˊ● ω ●ˋ", ""}
		context.Set(r, "user", u)
		h.Next(w, r)
	}
	return handlers.Handler{fn}
}

// Test Example for Handlers
func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test")
}

// Test JSON Example for Handlers
func testJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	u := context.Get(r, "user")
	json.NewEncoder(w).Encode(u)
}

// Test Error Example for Handlers
func testError(w http.ResponseWriter, r *http.Request) {
	error.ResError(w, error.CrazyError)
}

// Test Error Example for Handlers
func postJSON(w http.ResponseWriter, r *http.Request) {
	u := context.Get(r, "body").(*user)
	fmt.Println(u.Id)
	fmt.Println(u.FirstName)
	fmt.Println(u.LastName)
}

// Test HTML Example for Handlers
func getHTML(w http.ResponseWriter, r *http.Request) {
	u := context.Get(r, "user")
	server.Render(w, "index.html", u)
}

func main() {
	app := server.New()
	app.Use(handlers.LogH)
	app.Use(handlers.ErrorH)
	app.Get("/test", test)
	app.Get("/jsontest", testJSON, authH)
	app.Get("/jsonerror", testError)
	app.Post("/postjson", postJSON, handlers.BodyParseH(user{}))
	app.Get("/gethtml", getHTML, authH)
	http.ListenAndServe(":8080", app)
}
