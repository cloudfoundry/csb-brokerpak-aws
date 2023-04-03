package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type DynamoDBNamespaceService struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	Prefix          string `mapstructure:"prefix"`
}

func Read() (DynamoDBNamespaceService, error) {
	app, err := cfenv.Current()
	if err != nil {
		return DynamoDBNamespaceService{}, fmt.Errorf("error reading app env: %w", err)
	}

	if svs, err := app.Services.WithTag("namespace"); err == nil {
		return parseCredentials(svs[0].Credentials)
	}
	if svs, err := app.Services.WithLabel("aws-dynamodb"); err == nil {
		return parseCredentials(svs[0].Credentials)
	}
	return DynamoDBNamespaceService{}, fmt.Errorf("could not find service with tag 'namespace' or label 'aws-dynamodb'")
}

func parseCredentials(input map[string]any) (DynamoDBNamespaceService, error) {
	var r DynamoDBNamespaceService
	if err := mapstructure.Decode(input, &r); err != nil {
		return DynamoDBNamespaceService{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.AccessKeyID == "" || r.SecretAccessKey == "" || r.Region == "" || r.Prefix == "" {
		return DynamoDBNamespaceService{}, fmt.Errorf("parsed credentials are not valid")
	}

	return r, nil
}
