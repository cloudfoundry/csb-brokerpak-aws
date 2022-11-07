package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"

	"postgresqlapp/internal/connector"
)

const (
	tableName     = "test"
	keyColumn     = "keyname"
	valueColumn   = "valuedata"
	tlsQueryParam = "tls"
	schemaKey     = "schema"
)

func App(conn *connector.Connector) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods(http.MethodHead, http.MethodGet)
	r.Handle("/{schema}", checkSchemaMiddleware(http.HandlerFunc(handleCreateSchema(conn)))).Methods(http.MethodPut)
	r.Handle("/{schema}", checkSchemaMiddleware(http.HandlerFunc(handleDropSchema(conn)))).Methods(http.MethodDelete)
	r.Handle("/{schema}/{key}", checkSchemaMiddleware(http.HandlerFunc(handleSet(conn)))).Methods(http.MethodPut)
	r.Handle("/{schema}/{key}", checkSchemaMiddleware(http.HandlerFunc(handleGet(conn)))).Methods(http.MethodGet)

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func checkSchemaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		schema, err := schemaName(r)
		if err != nil {
			fail(w, http.StatusInternalServerError, "Schema name error: %s\n", err)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), schemaKey, schema)))
	})
}

func schemaName(r *http.Request) (string, error) {
	schema, ok := mux.Vars(r)["schema"]

	switch {
	case !ok:
		return "", fmt.Errorf("schema missing")
	case len(schema) > 50:
		return "", fmt.Errorf("schema name too long")
	case len(schema) == 0:
		return "", fmt.Errorf("schema name cannot be zero length")
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
