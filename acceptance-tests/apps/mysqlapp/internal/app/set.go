package app

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"mysqlapp/internal/connector"

	"github.com/gorilla/mux"
)

func handleSet(conn *connector.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")

		db, err := conn.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
		if err != nil {
			fail(w, http.StatusInternalServerError, "error connecting to database: %s", err)
		}

		_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`, tableName, keyColumn, valueColumn))
		if err != nil {
			fail(w, http.StatusInternalServerError, "failed to create test table: %s", err)
			return
		}

		key, ok := mux.Vars(r)["key"]
		if !ok {
			fail(w, http.StatusBadRequest, "key missing")
			return
		}

		rawValue, err := io.ReadAll(r.Body)
		if err != nil {
			fail(w, http.StatusBadRequest, "error parsing value: %s", err)
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`INSERT INTO %s (%s, %s) VALUES (?, ?)`, tableName, keyColumn, valueColumn))
		if err != nil {
			fail(w, http.StatusInternalServerError, "error preparing statement: %s", err)
			return
		}
		defer stmt.Close()

		if _, err := stmt.Exec(key, string(rawValue)); err != nil {
			fail(w, http.StatusBadRequest, "error inserting values: %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Key %q set to value %q.", key, string(rawValue))
	}
}
