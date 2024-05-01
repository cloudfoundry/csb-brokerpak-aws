package app

import (
	"io"
	"log"
	"net/http"
	"redisapp/internal/credentials"
)

func handleSet(creds credentials.Credentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")

		key := r.PathValue("key")
		if key == "" {
			fail(w, http.StatusBadRequest, "url parameter 'key' is required")
			return
		}

		rawValue, err := io.ReadAll(r.Body)
		if err != nil {
			fail(w, http.StatusBadRequest, "error parsing value from body: %s", err)
			return
		}

		value := string(rawValue)
		if err := creds.Client().Set(r.Context(), key, value, 0).Err(); err != nil {
			fail(w, http.StatusFailedDependency, "failed to set value in Redis: %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Key %q set to value %q.", key, value)
	}
}
