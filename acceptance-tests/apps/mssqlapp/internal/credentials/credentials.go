package credentials

import (
	"fmt"
	"log"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	URI      string `mapstructure:"uri"`
	Server   string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"name"`
}

type LegacyConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	URI      string `mapstructure:"uri"`
	Server   string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
}

func Read() (*Connector, error) {
	app, err := cfenv.Current()
	if err != nil {
		return nil, fmt.Errorf("error reading app env: %w", err)
	}
	if svs, err := app.Services.WithTag("mssql"); err == nil {
		log.Println("found tag: mssql")
		return readCredentials(svs)
	}

	if svs, err := app.Services.WithLabel("aws-rds-sqlserver"); err == nil {
		log.Println("found label: aws-rds-sqlserver")
		return readLegacyCredentials(svs)
	}

	return nil, fmt.Errorf("error reading MSSQL service details")
}

func readCredentials(svs []cfenv.Service) (*Connector, error) {
	var c Config
	if err := mapstructure.Decode(svs[0].Credentials, &c); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	connector := NewConnector(
		c.Server,
		c.Username,
		c.Password,
		c.Database,
		c.Port,
	)
	return connector, nil
}

func readLegacyCredentials(svs []cfenv.Service) (*Connector, error) {
	var c LegacyConfig
	if err := mapstructure.Decode(svs[0].Credentials, &c); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	connector := NewConnector(
		c.Server,
		c.Username,
		c.Password,
		c.Database,
		c.Port,
	)
	return connector, nil
}
