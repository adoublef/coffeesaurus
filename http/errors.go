package http

import (
	"log"
	"net/http"
)

// Error replies to the request
func Error(w http.ResponseWriter, err error, code int) {
	// TODO -- option to pretty print error
	log.Printf("Error: %v\n\n", err)
	http.Error(w, http.StatusText(code), code)
}
