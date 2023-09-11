package app

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/go-chi/chi/v5"

	"mssqlapp/internal/credentials"
)

const (
	tableName   = "test"
	keyColumn   = "keyname"
	valueColumn = "valuedata"

	tlsQueryParam = "tls"
)

func App(connector *credentials.Connector) http.Handler {
	r := chi.NewRouter()
	r.Head("/", aliveness)
	r.Put("/{schema}", handleCreateSchema(connector))
	r.Post("/{schema}", handleFillDatabase(connector))
	r.Delete("/{schema}", handleDropSchema(connector))
	r.Put("/{schema}/{key}", handleSet(connector))
	r.Get("/{schema}/{key}", handleGet(connector))

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func schemaName(r *http.Request) (string, error) {
	schema := chi.URLParam(r, "schema")

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
