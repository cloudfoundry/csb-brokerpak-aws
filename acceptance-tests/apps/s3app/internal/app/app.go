package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"s3app/internal/credentials"
)

func App(client *credentials.Client) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		filename := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodPut:
			handleUpload(w, r, filename, client)
		case http.MethodGet:
			handleDownload(w, r, filename, client)
		case http.MethodDelete:
			handleDelete(w, r, filename, client)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
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
