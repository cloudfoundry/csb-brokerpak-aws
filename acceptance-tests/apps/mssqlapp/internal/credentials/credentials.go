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

func Read() (string, error) {
	app, err := cfenv.Current()
	if err != nil {
		return "", fmt.Errorf("error reading app env: %w", err)
	}
	if svs, err := app.Services.WithTag("mssql"); err == nil {
		log.Println("found tag: mssql")
		return readService(svs)
	}

	return "", fmt.Errorf("error reading MSSQL service details")
}

func readService(svs []cfenv.Service) (string, error) {
	var c Config
	if err := mapstructure.Decode(svs[0].Credentials, &c); err != nil {
		return "", fmt.Errorf("failed to decode credentials: %w", err)
	}

	enc := NewEncoder(
		c.Server,
		c.Username,
		c.Password,
		c.Database,
		"true",
		c.Port,
	)
	return enc.Encode(), nil
}
