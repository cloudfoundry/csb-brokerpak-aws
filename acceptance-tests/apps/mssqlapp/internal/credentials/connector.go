package credentials

import (
	"database/sql"
	"fmt"
	"log"
)

type Connector struct {
	server   string
	username string
	password string
	database string
	port     int
}

func NewConnector(
	server,
	username,
	password,
	database string,
	port int,
) *Connector {
	return &Connector{
		server:   server,
		username: username,
		password: password,
		port:     port,
		database: database,
	}
}

func (c *Connector) Connect(enc *Encoder) (*sql.DB, error) {

	stringConn := enc.Encode()
	db, err := sql.Open("sqlserver", stringConn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database. Configuration: %q: %w", stringConn, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database. Configuration: %q: %w", stringConn, err)
	}

	return db, nil
}

func (c *Connector) WithTLS(tls string) *Encoder {
	switch tls {
	case "disable":
		return NewEncoder(c.server, c.username, c.password, c.database, "disable", c.port)
	case "", "true":
		return NewEncoder(c.server, c.username, c.password, c.database, "true", c.port)
	default:
		log.Fatalf("tls value not implemented: %s", tls)
	}
	return nil
}
