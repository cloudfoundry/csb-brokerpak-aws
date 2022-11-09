package connector

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	URI        string `mapstructure:"uri"`
	Parameters map[string]any
}

func New() (*Connector, error) {
	svc, err := findService()
	if err != nil {
		return nil, fmt.Errorf("error finding service binding details: %w", err)
	}

	c := Connector{Parameters: make(map[string]any)}
	if err := mapstructure.Decode(svc.Credentials, &c); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if c.URI == "" {
		return nil, fmt.Errorf("parsed credentials are not valid")
	}

	return &c, nil
}

func (c *Connector) Connect(opts ...Option) (*sql.DB, error) {
	if err := c.withDefaults(opts...); err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", c.generateURI())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to connect to database", err)
	}

	db.SetMaxIdleConns(0)

	if err := db.Ping(); err != nil {
		return db, fmt.Errorf("failed to verify the connection to the database is still alive")
	}

	return db, nil
}

func (c *Connector) generateURI() string {
	if len(c.Parameters) == 0 {
		return c.URI
	}

	v := url.Values{}
	for key, value := range c.Parameters {
		v.Set(key, fmt.Sprintf("%v", value))
	}

	var s strings.Builder
	s.WriteString(c.URI)
	s.WriteString("?")
	s.WriteString(v.Encode())
	return s.String()
}

func (c *Connector) withDefaults(opts ...Option) error {
	opts = append([]Option{WithTLS("require")}, opts...)
	for _, o := range opts {
		if err := o(c); err != nil {
			return err
		}
	}
	return nil

}

func findService() (cfenv.Service, error) {
	app, err := cfenv.Current()
	if err != nil {
		return cfenv.Service{}, fmt.Errorf("error reading app env: %w", err)
	}

	for _, f := range []func() ([]cfenv.Service, error){
		func() ([]cfenv.Service, error) { return app.Services.WithTag("postgresql") },
		func() ([]cfenv.Service, error) { return app.Services.WithLabel("aws-rds-postgres") },
	} {
		svs, err := f()
		if err == nil && len(svs) > 0 {
			return svs[0], nil
		}
	}

	return cfenv.Service{}, fmt.Errorf("unable to find credentials for PostgreSQL")
}

type Option func(*Connector) error

func WithTLS(tls string) Option {
	return func(c *Connector) error {
		switch tls {
		case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
			c.Parameters["sslmode"] = tls
		case "":
			c.Parameters["sslmode"] = "require"
		default:
			return fmt.Errorf("invalid tls value: %s", tls)
		}
		return nil
	}
}
