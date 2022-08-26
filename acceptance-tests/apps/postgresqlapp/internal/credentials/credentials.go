package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

func Read() (string, error) {
	svc, err := findService()
	if err != nil {
		return "", fmt.Errorf("error finding service binding details: %w", err)
	}

	var m struct {
		URI string `mapstructure:"uri"`
	}

	if err := mapstructure.Decode(svc.Credentials, &m); err != nil {
		return "", fmt.Errorf("failed to decode credentials: %w", err)
	}

	if m.URI == "" {
		return "", fmt.Errorf("parsed credentials are not valid")
	}

	return m.URI, nil
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
