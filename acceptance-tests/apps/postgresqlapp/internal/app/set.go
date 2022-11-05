package app

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"postgresqlapp/internal/connector"

	"github.com/gorilla/mux"
)

func handleSet(conn *connector.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")

		db, err := conn.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
		if err != nil {
			fail(w, http.StatusInternalServerError, "failed to connect to database: %e", err)
		}
		defer db.Close()

		schema := r.Context().Value(schemaKey)
		key, ok := mux.Vars(r)["key"]
		if !ok {
			fail(w, http.StatusBadRequest, "Key missing.")
			return
		}

		rawValue, err := io.ReadAll(r.Body)
		if err != nil {
			fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`INSERT INTO %s.%s (%s, %s) VALUES ($1, $2)`, schema, tableName, keyColumn, valueColumn))
		if err != nil {
			fail(w, http.StatusInternalServerError, "Error preparing statement: %s", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(key, string(rawValue))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error inserting values: %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Key %q set to value %q.", key, string(rawValue))
	}
}
