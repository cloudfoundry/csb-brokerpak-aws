package app

import (
	"fmt"
	"log"
	"net/http"

	"mssqlapp/internal/credentials"
)

func handleDropSchema(connector *credentials.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling drop schema.")
		db, err := connector.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
		if err != nil {
			fail(w, http.StatusInternalServerError, "failed to connect to database: %s", err)
			return
		}

		schema, err := schemaName(r)
		if err != nil {
			fail(w, http.StatusInternalServerError, "schema name error: %s", err)
			return
		}

		if _, err = db.Exec(fmt.Sprintf(`DROP TABLE %s.%s`, schema, tableName)); err != nil {
			fail(w, http.StatusBadRequest, "error dropping table: %s", err)
			return
		}

		if _, err = db.Exec(fmt.Sprintf(`DROP SCHEMA %s`, schema)); err != nil {
			fail(w, http.StatusBadRequest, "error dropping schema: %s", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("Schema %q dropped", schema)
	}
}
