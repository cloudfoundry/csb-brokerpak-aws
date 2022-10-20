package connector

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type option func(*Connector, *mysql.Config) error

func (c *Connector) Connect(opts ...option) (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = c.Host
	cfg.User = c.Username
	cfg.Passwd = c.Password
	cfg.DBName = c.Database
	withDefaults(opts...)(c, cfg)

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

func withOptions(opts ...option) option {
	return func(conn *Connector, cfg *mysql.Config) error {
		for _, o := range opts {
			if err := o(conn, cfg); err != nil {
				return err
			}
		}
		return nil
	}
}

func withDefaults(opts ...option) option {
	return withOptions(append([]option{withTLS()}, opts...)...)
}

func withTLS() option {
	return func(_ *Connector, cfg *mysql.Config) error {
		cfg.TLSConfig = "true"
		return nil
	}
}
