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

		schema, err := schemaName(r)
		if err != nil {
			fail(w, http.StatusBadRequest, "schema name error: %s", err)
			return
		}

		db, err := conn.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
		if err != nil {
			fail(w, http.StatusInternalServerError, "failed to connect to database: %s", err)
		}
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf(`DROP SCHEMA %s CASCADE`, schema))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error dropping schema: %s", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("Schema %q dropped", schema)
	}
}
