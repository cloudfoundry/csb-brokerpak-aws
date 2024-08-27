package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type Credentials map[string]Credential

func Read() (Credentials, error) {
	app, err := cfenv.Current()
	if err != nil {
		return Credentials{}, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("sqs")
	if err != nil {
		return Credentials{}, fmt.Errorf("error reading SQS service details")
	}

	creds := make(Credentials)
	for i, s := range svs {
		var r Credential
		if err := mapstructure.Decode(s.Credentials, &r); err != nil {
			return Credentials{}, fmt.Errorf("failed to decode credentials for binding %q (%d): %w", s.Name, i, err)
		}

		if err := r.validate(); err != nil {
			return Credentials{}, fmt.Errorf("validation error for binding %q (%d): %w", s.Name, i, err)
		}

		creds[s.Name] = r
	}

	return creds, nil
}
