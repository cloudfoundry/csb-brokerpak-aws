package app

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"mssqlapp/internal/credentials"
)

func handleSet(connector *credentials.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")
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

		key := r.PathValue("key")
		if key == "" {
			fail(w, http.StatusBadRequest, "key must be supplied")
			return
		}

		rawValue, err := io.ReadAll(r.Body)
		if err != nil {
			fail(w, http.StatusBadRequest, "error parsing value: %s", err)
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`INSERT INTO %s.%s (%s, %s) VALUES (@p1, @p2)`, schema, tableName, keyColumn, valueColumn))
		if err != nil {
			fail(w, http.StatusInternalServerError, "error preparing statement: %s", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(key, string(rawValue))
		if err != nil {
			fail(w, http.StatusBadRequest, "failed to insert value: %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Key %q set to value %q in schema %q.", key, string(rawValue), schema)
	}
}
