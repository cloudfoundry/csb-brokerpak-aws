package connector

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Option func(*Connector, *mysql.Config) error

func (c *Connector) Connect(opts ...Option) (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = c.Host
	cfg.User = c.Username
	cfg.Passwd = c.Password
	cfg.DBName = c.Database

	if err := withDefaults(opts...)(c, cfg); err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

func withOptions(opts ...Option) Option {
	return func(conn *Connector, cfg *mysql.Config) error {
		for _, o := range opts {
			if err := o(conn, cfg); err != nil {
				return err
			}
		}
		return nil
	}
}

func withDefaults(opts ...Option) Option {
	return withOptions(append([]Option{WithTLS("true")}, opts...)...)
}

func WithTLS(tls string) Option {
	return func(_ *Connector, cfg *mysql.Config) error {
		switch tls {
		case "false", "true", "skip-verify", "preferred":
			cfg.TLSConfig = tls
		case "":
			cfg.TLSConfig = "true"
		default:
			return fmt.Errorf("invalid tls value: %s", tls)
		}
		return nil
	}
}
