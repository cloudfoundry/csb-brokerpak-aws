package app

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"mssqlapp/internal/credentials"

	_ "github.com/denisenkom/go-mssqldb"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"

	tlsQueryParam = "tls"
)

func App(connector *credentials.Connector) http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /", aliveness)
	r.HandleFunc("PUT /{schema}", handleCreateSchema(connector))
	r.HandleFunc("POST /{schema}", handleFillDatabase(connector))
	r.HandleFunc("DELETE /{schema}", handleDropSchema(connector))
	r.HandleFunc("PUT /{schema}/{key}", handleSet(connector))
	r.HandleFunc("GET /{schema}/{key}", handleGet(connector))

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func schemaName(r *http.Request) (string, error) {
	schema := r.PathValue("schema")

	switch {
	case schema == "":
		return "", fmt.Errorf("schema name must be supplied")
	case len(schema) > 50:
		return "", fmt.Errorf("schema name too long")
	case !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(schema):
		return "", fmt.Errorf("schema name contains invalid characters")
	default:
		return schema, nil
	}
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
