package app

import (
	"fmt"
	"log"
	"net/http"
	"redisapp/internal/credentials"

	"github.com/go-chi/chi/v5"
)

func App(creds credentials.Credentials) http.Handler {
	r := chi.NewRouter()

	r.Head("/", aliveness)
	r.Put("/primary/{key}", handleSet(creds))
	r.Get("/primary/{key}", handleGet(creds, primaryNode))
	r.Get("/replica/{key}", handleGet(creds, replicaNode))

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
