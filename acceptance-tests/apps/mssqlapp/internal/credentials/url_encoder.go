package credentials

import (
	"fmt"
	"net/url"
)

const (
	queryParamDatabaseKey = "database"
	queryParamEncryptKey  = "encrypt"

	queryParamTrustServerCertificate = "TrustServerCertificate"
	queryParamHostNameInCertificate  = "HostNameInCertificate"
)

type Encoder struct {
	server      string
	username    string
	password    string
	port        int
	queryParams map[string]string
}

func NewEncoder(
	server,
	username,
	password,
	database,
	encrypt string,
	port int,
) *Encoder {
	queryParams := map[string]string{queryParamDatabaseKey: database, queryParamEncryptKey: encrypt}

	if encrypt == "true" {
		queryParams[queryParamTrustServerCertificate] = "false"
		queryParams[queryParamHostNameInCertificate] = server
	}

	return &Encoder{
		server:      server,
		username:    username,
		password:    password,
		port:        port,
		queryParams: queryParams,
	}
}

func (e *Encoder) Encode() string {
	u := createURL(e.server, e.username, e.password, e.port)
	u.RawQuery = createQueryParams(e.queryParams).Encode()

	return u.String()
}

func (e *Encoder) withEncrypt() *Encoder {
	e.queryParams[queryParamEncryptKey] = "true"
	e.queryParams[queryParamTrustServerCertificate] = "false"
	e.queryParams[queryParamHostNameInCertificate] = e.server
	return e
}

func (e *Encoder) withoutEncrypt() *Encoder {
	e.queryParams[queryParamEncryptKey] = "disable"
	e.queryParams[queryParamTrustServerCertificate] = "true"
	e.queryParams[queryParamHostNameInCertificate] = ""
	return e
}

func createQueryParams(params map[string]string) url.Values {
	q := url.Values{}
	for key, value := range params {
		if value != "" {
			q.Add(key, value)
		}
	}
	return q
}

func createURL(server, username, password string, port int) url.URL {
	const scheme = "sqlserver"
	u := url.URL{
		Scheme: scheme,
		User:   url.UserPassword(username, password),
		Host:   fmt.Sprintf("%s:%d", server, port),
	}

	return u
}
