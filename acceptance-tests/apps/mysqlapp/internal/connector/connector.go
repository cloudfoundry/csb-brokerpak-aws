package connector

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	Host     string `mapstructure:"hostname"`
	Database string `mapstructure:"name"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
}

func New() (*Connector, error) {
	app, err := cfenv.Current()
	if err != nil {
		return nil, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("mysql")
	if err != nil {
		return nil, fmt.Errorf("error reading MySQL service details")
	}

	var c Connector
	if err := mapstructure.Decode(svs[0].Credentials, &c); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if err := c.Valid(); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Connector) Valid() error {
	switch {
	case c.Host == "":
		return fmt.Errorf("missing hostname")
	case c.Username == "":
		return fmt.Errorf("missing username")
	case c.Password == "":
		return fmt.Errorf("missing password")
	case c.Database == "":
		return fmt.Errorf("missing database name")
	case c.Port == 0:
		return fmt.Errorf("missing port")
	}

	return nil
}
