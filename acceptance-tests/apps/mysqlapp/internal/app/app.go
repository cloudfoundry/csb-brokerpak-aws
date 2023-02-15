package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"mysqlapp/internal/connector"
)

const (
	tableName     = "test"
	keyColumn     = "keyname"
	valueColumn   = "valuedata"
	tlsQueryParam = "tls"
)

func App(conn *connector.Connector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodGet:
			handleGet(w, r, key, conn)
		case http.MethodPut:
			handleSet(w, r, key, conn)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
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
