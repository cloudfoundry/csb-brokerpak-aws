package connector

import (
	"fmt"
	"net/url"
)

type URLEncoder interface {
	String() string
}

type awsEncoder struct {
	server   string
	username string
	password string
	database string
	encrypt  string
	port     int
}

func NewAWSEncoder(
	server,
	username,
	password,
	database,
	encrypt string,
	port int,
) *awsEncoder {
	return &awsEncoder{
		server:   server,
		username: username,
		password: password,
		database: database,
		encrypt:  encrypt,
		port:     port,
	}
}

func (a *awsEncoder) String() string {
	query := url.Values{}
	// query.Add("database", a.database)
	query.Add("encrypt", a.encrypt)

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(a.username, a.password),
		Host:     fmt.Sprintf("%s:%d", a.server, a.port),
		RawQuery: query.Encode(),
	}

	return u.String()
}

type azureEncoder struct {
	server   string
	username string
	password string
	database string
	encrypt  string
	port     int
}

func NewAzureEncoder(
	server,
	username,
	password,
	database,
	encrypt string,
	port int,
) *azureEncoder {
	return &azureEncoder{
		server:   server,
		username: username,
		password: password,
		database: database,
		encrypt:  encrypt,
		port:     port,
	}
}

func (a *azureEncoder) String() string {
	query := url.Values{}
	query.Add("database", a.database)
	query.Add("encrypt", a.encrypt)

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(a.username, a.password),
		Host:     fmt.Sprintf("%s:%d", a.server, a.port),
		RawQuery: query.Encode(),
	}

	return u.String()
}
