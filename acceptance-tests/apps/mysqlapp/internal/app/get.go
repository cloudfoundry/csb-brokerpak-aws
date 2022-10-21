package app

import (
	"fmt"
	"log"
	"net/http"

	"mysqlapp/internal/connector"

	"github.com/gorilla/mux"
)

func handleGet(conn *connector.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get.")

		db, err := conn.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
		if err != nil {
			fail(w, http.StatusInternalServerError, "error connecting to database: %s", err)
		}

		key, ok := mux.Vars(r)["key"]
		if !ok {
			fail(w, http.StatusBadRequest, "key missing")
			return
		}

		stmt, err := db.Prepare(fmt.Sprintf(`SELECT %s from %s WHERE %s = ?`, valueColumn, tableName, keyColumn))
		if err != nil {
			fail(w, http.StatusInternalServerError, "error preparing statement: %s", err)
			return
		}
		defer stmt.Close()

		rows, err := stmt.Query(key)
		if err != nil {
			fail(w, http.StatusNotFound, "error selecting value: %s", err)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			fail(w, http.StatusNotFound, "error finding value: %s", err)
			return
		}

		var value string
		if err := rows.Scan(&value); err != nil {
			fail(w, http.StatusNotFound, "error retrieving value: %s", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(value)); err != nil {
			log.Printf("Error writing value: %s", err)
			return
		}

		log.Printf("Value %q retrived from key %q.", value, key)
	}
}
