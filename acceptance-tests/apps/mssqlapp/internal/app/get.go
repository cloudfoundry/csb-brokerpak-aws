package app

import (
	"fmt"
	"log"
	"net/http"

	"mssqlapp/internal/credentials"
)

func handleGet(connector *credentials.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get.")
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

		stmt, err := db.Prepare(fmt.Sprintf(`SELECT %s from %s.%s WHERE %s = @p1`, valueColumn, schema, tableName, keyColumn))
		if err != nil {
			fail(w, http.StatusInternalServerError, "error preparing statement: %s", err)
			return
		}
		defer stmt.Close()

		rows, err := stmt.Query(key)
		if err != nil {
			fail(w, http.StatusNotFound, "failed to select value for key: %s", key)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			fail(w, http.StatusNotFound, "failed to find value for key: %s", key)
			return
		}

		var value string
		if err := rows.Scan(&value); err != nil {
			fail(w, http.StatusNotFound, "failed to retrieve value for key: %s", key)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write([]byte(value))

		if err != nil {
			log.Printf("Error writing value: %s", err)
			return
		}

		log.Printf("Value %q retrived from key %q in schema %s.", value, key, schema)
	}
}
