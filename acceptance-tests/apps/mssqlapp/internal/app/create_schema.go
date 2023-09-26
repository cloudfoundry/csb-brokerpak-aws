package app

import (
	"fmt"
	"log"
	"net/http"

	"mssqlapp/internal/credentials"
)

func handleCreateSchema(connector *credentials.Connector) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling create schema.")
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

		statement := fmt.Sprintf(`CREATE SCHEMA %s`, schema)
		switch r.URL.Query().Get("dbo") {
		case "true":
			statement = statement + " AUTHORIZATION dbo"
		case "":
			break // default value " AUTHORIZATION connected user"
		case "false":
		default:
			fail(w, http.StatusBadRequest, "invalid value for dbo")
			return
		}

		if _, err = db.Exec(statement); err != nil {
			fail(w, http.StatusBadRequest, "failed to create schema: %s", err)
			return
		}

		if _, err = db.Exec(fmt.Sprintf(`CREATE TABLE %s.%s (%s VARCHAR(255) NOT NULL, %s VARCHAR(max) NOT NULL)`, schema, tableName, keyColumn, valueColumn)); err != nil {
			fail(w, http.StatusBadRequest, "error creating table: %s", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Schema %q created", schema)
	}
}
