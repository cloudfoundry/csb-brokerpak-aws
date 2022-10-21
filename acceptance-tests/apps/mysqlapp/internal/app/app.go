package app

import (
	"fmt"
	"log"
	"net/http"

	"mysqlapp/internal/connector"

	"github.com/gorilla/mux"
)

const (
	tableName     = "test"
	keyColumn     = "keyname"
	valueColumn   = "valuedata"
	tlsQueryParam = "tls"
)

func App(conn *connector.Connector) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")
	r.HandleFunc("/{key}", handleSet(conn)).Methods("PUT")
	r.HandleFunc("/{key}", handleGet(conn)).Methods("GET")

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
