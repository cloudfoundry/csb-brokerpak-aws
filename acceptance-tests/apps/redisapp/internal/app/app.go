package app

import (
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func App(options *redis.Options) *mux.Router {
	client := redis.NewClient(options)
	r := mux.NewRouter()

	r.HandleFunc("/", aliveness).Methods(http.MethodHead, http.MethodGet)
	r.HandleFunc("/{key}", handleSet(client)).Methods(http.MethodPut)
	r.HandleFunc("/{key}", handleGet(client)).Methods(http.MethodGet)

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}
