// Package app provides functionality for receiving messages from an SQS queue.
package app

import (
	"log"
	"net/http"

	"sqsapp/internal/credentials"
)

func App(creds credentials.Credentials) http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /", aliveness)
	r.HandleFunc("GET /retrieve_and_delete/{binding_name}", writeResponse(handleRetrieveAndDelete(creds)))
	r.HandleFunc("GET /retrieve/{binding_name}", writeResponse(handleRetrieve(creds)))
	r.HandleFunc("POST /send/{binding_name}", writeResponse(handleSend(creds)))

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

// writeResponse allows handler functions to simply return an HTTP code and a message
// avoiding repeated boilerplate code for dealing with the http.ResponseWriter
func writeResponse(h func(r *http.Request) (int, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, msg := h(r)
		switch code {
		case http.StatusOK:
			w.WriteHeader(code)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(msg))
		case http.StatusNoContent:
			w.WriteHeader(code)
		default:
			http.Error(w, msg, code)
		}
	}
}
