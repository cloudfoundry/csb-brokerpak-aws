package app

import (
	"io"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

func handleSet(w http.ResponseWriter, r *http.Request, key string, client *redis.Client) {
	log.Println("Handling set.")

	rawValue, err := io.ReadAll(r.Body)
	if err != nil {
		fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
		return
	}

	value := string(rawValue)
	if err := client.Set(r.Context(), key, value, 0).Err(); err != nil {
		fail(w, http.StatusFailedDependency, "Failed to set value: %s", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Key %q set to value %q.", key, value)
}
