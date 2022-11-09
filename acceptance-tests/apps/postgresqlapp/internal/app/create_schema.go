package app

import (
	"fmt"
	"log"
	"net/http"

	"postgresqlapp/internal/connector"
)

func handleCreateSchema(conn *connector.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling create schema.")

		db, err := conn.Connect(connector.WithTLS(r.URL.Query().Get(tlsQueryParam)))
		if err != nil {
			fail(w, http.StatusInternalServerError, "failed to connect to database: %e", err)
		}
		defer db.Close()

		schema := r.Context().Value(schemaKey)
		_, err = db.Exec(fmt.Sprintf(`CREATE SCHEMA %s`, schema))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error creating schema: %s", err)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`GRANT ALL ON SCHEMA %s TO PUBLIC`, schema))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error granting schema permissions: %s", err)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (%s VARCHAR(255) NOT NULL, %s VARCHAR(255) NOT NULL)`, schema, tableName, keyColumn, valueColumn))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error creating table: %s", err)
			return
		}

		_, err = db.Exec(fmt.Sprintf(`GRANT ALL ON TABLE %s.%s TO PUBLIC`, schema, tableName))
		if err != nil {
			fail(w, http.StatusBadRequest, "Error granting table permissions: %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Schema %q created", schema)
	}
}
