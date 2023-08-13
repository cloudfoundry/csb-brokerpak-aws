package connector

import (
	"fmt"
	"net/url"
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
	encrypt,
	iaas string,
	port int,
) *Encoder {
	var queryParams map[string]string
	switch iaas {
	case AWS:
		queryParams = map[string]string{"encrypt": encrypt}
	case Azure:
		queryParams = map[string]string{"database": database, "encrypt": encrypt}
	}
	return &Encoder{
		server:      server,
		username:    username,
		password:    password,
		port:        port,
		queryParams: queryParams,
	}
}

func (b *Encoder) Encode() string {
	u := createURL(b.server, b.username, b.password, b.port)
	u.RawQuery = createQueryParams(b.queryParams).Encode()

	return u.String()
}

func createQueryParams(params map[string]string) url.Values {
	q := url.Values{}
	for key, value := range params {
		q.Add(key, value)
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
