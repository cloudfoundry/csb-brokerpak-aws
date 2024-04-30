package app

import (
	"fmt"
	"log"
	"net/http"
	"postgresqlapp/internal/connector"
	"regexp"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	tableName     = "test"
	keyColumn     = "keyname"
	valueColumn   = "valuedata"
	tlsQueryParam = "tls"
)

func App(conn *connector.Connector) http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("GET /aliveness", aliveness)
	r.HandleFunc("PUT /{schema}", handleCreateSchema(conn))
	r.HandleFunc("DELETE /{schema}", handleDropSchema(conn))
	r.HandleFunc("PUT /{schema}/{key}", handleSet(conn))
	r.HandleFunc("GET /{schema}/{key}", handleGet(conn))

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
		return "", fmt.Errorf("schema must be supplied")
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
