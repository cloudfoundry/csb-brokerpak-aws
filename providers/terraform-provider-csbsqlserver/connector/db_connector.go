package connector

import (
	"database/sql"
	"fmt"
)

type DBConnector struct {
	server   string
	username string
	password string
	database string
	encrypt  string
	port     int
	encoder  *Encoder
}

func NewDBConnector(server string, port int, username, password, database, encrypt string) *DBConnector {
	return &DBConnector{
		server:   server,
		username: username,
		password: password,
		database: database,
		encrypt:  encrypt,
		port:     port,
		encoder: NewEncoder(
			server,
			username,
			password,
			database,
			encrypt,
			port,
		),
	}
}

func (dbc *DBConnector) withConnection(callback func(*sql.DB) error) error {
	db, err := dbc.connect(dbc.encoder.Encode())
	if err != nil {
		return err
	}

	return callback(db)
}

func (dbc *DBConnector) withMasterDBConnection(callback func(*sql.DB) error) error {
	db, err := dbc.connect(dbc.encoder.EncodeWithoutDB())
	if err != nil {
		return err
	}

	return callback(db)
}

func (dbc *DBConnector) withTransaction(callback func(tx *sql.Tx) error) error {
	return dbc.withConnection(func(db *sql.DB) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		if err := callback(tx); err != nil {
			return err
		}

		return tx.Commit()
	})
}

func (dbc *DBConnector) connect(url string) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", url)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database on %q port %d with user %q: %w", dbc.server, dbc.port, dbc.username, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database on %q port %d with user %q: %w", dbc.server, dbc.port, dbc.username, err)
	}

	return db, nil
}
