package app

import (
	"fmt"
	"log"
	"net/http"
	"redisapp/internal/credentials"
)

func App(creds credentials.Credentials) http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /", aliveness)
	r.HandleFunc("PUT /primary/{key}", handleSet(creds))
	r.HandleFunc("GET /primary/{key}", handleGet(creds, primaryNode))
	r.HandleFunc("GET /replica/{key}", handleGet(creds, replicaNode))

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
