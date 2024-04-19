package app

import (
	"log"
	"net/http"
	"redisapp/internal/credentials"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

const (
	primaryNode = true
	replicaNode = false
)

func handleGet(creds credentials.Credentials, primary bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling get (primary=%t).", primary)

		key := chi.URLParam(r, "key")
		if key == "" {
			fail(w, http.StatusBadRequest, "url parameter 'key' is required")
			return
		}

		c, err := client(creds, primary)
		if err != nil {
			fail(w, http.StatusFailedDependency, "could not create client: %s", err)
			return
		}

		value, err := c.Get(r.Context(), key).Result()
		if err != nil {
			fail(w, http.StatusNotFound, "error retrieving value from Redis: %s", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write([]byte(value))

		if err != nil {
			log.Printf("Error writing value: %s", err)
			return
		}

		log.Printf("Value %q retrieved from key %q.", value, key)
	}
}

func client(creds credentials.Credentials, primary bool) (*redis.Client, error) {
	switch primary {
	case primaryNode:
		return creds.Client(), nil
	default:
		return creds.ReaderClient()
	}
}
