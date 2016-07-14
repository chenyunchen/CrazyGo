package error

import (
	"encoding/json"
	"net/http"
)

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
