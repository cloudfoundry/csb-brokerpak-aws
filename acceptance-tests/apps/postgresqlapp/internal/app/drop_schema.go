package app

import (
	"fmt"
	"log"
	"net/http"

	"postgresqlapp/internal/connector"
)

func handleDropSchema(conn *connector.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling drop schema.")

		db, err := conn.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
		if err != nil {
			fail(w, http.StatusInternalServerError, "failed to connect to database: %e", err)
		}
		defer db.Close()

		schema := r.Context().Value(schemaKey)
		_, err = db.Exec(fmt.Sprintf(`DROP SCHEMA %s CASCADE`, schema))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error dropping schema: %s", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("Schema %q dropped", schema)
	}
}
