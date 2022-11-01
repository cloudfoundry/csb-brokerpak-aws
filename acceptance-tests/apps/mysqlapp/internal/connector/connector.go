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

type LegacyConnector struct {
	Host     string `mapstructure:"hostname"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
}

func New() (*Connector, error) {
	serviceTag, serviceCred, err := findService()
	if err != nil {
		return nil, fmt.Errorf("error finding service binding details: %w", err)
	}

	switch serviceTag {
	case "mysql":
		return ReadCSBMySQL(serviceCred)
	case "aws-rds-mysql":
		return ReadLegacyMySQL(serviceCred)
	}

	return nil, fmt.Errorf("unable to find credentials for MySQL")
}

func findService() (string, cfenv.Service, error) {
	app, err := cfenv.Current()
	if err != nil {
		return "", cfenv.Service{}, fmt.Errorf("error reading app env: %w", err)
	}

	for _, f := range []func() (string, []cfenv.Service, error){
		func() (string, []cfenv.Service, error) {
			serviceTag := "mysql"
			srv, err := app.Services.WithTag(serviceTag)
			return serviceTag, srv, err
		},
		func() (string, []cfenv.Service, error) {
			serviceLabel := "aws-rds-mysql"
			srv, err := app.Services.WithLabel(serviceLabel)
			return serviceLabel, srv, err
		},
	} {
		serviceType, svs, err := f()
		if err == nil && len(svs) > 0 {
			return serviceType, svs[0], nil
		}
	}

	return "", cfenv.Service{}, fmt.Errorf("unable to find credentials for MySQL")
}

func ReadCSBMySQL(svc cfenv.Service) (*Connector, error) {
	var c Connector
	if err := mapstructure.Decode(svc.Credentials, &c); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if err := c.Valid(); err != nil {
		return nil, err
	}

	return &c, nil
}

func ReadLegacyMySQL(svc cfenv.Service) (*Connector, error) {
	var c LegacyConnector
	if err := mapstructure.Decode(svc.Credentials, &c); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	cc := Connector{
		Host:     c.Host,
		Database: c.Database,
		Username: c.Username,
		Password: c.Password,
		Port:     c.Port,
	}
	if err := cc.Valid(); err != nil {
		return nil, err
	}

	return &cc, nil
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
