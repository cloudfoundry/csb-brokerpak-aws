// Package app provides functionality for receiving messages from an SQS queue.
package app

import (
	"fmt"
	"log"
	"net/http"

	"sqsapp/internal/credentials"
)

func App(creds credentials.Credentials) http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /", aliveness)
	r.HandleFunc("GET /receive", handleReceive(creds))
	r.HandleFunc("POST /send", handleSend(creds))

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
