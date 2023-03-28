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
	svs, err := app.Services.WithTag("namespace")
	if err != nil {
		return DynamoDBNamespaceService{}, fmt.Errorf("error reading DynamoDB Namespace service details: %w", err)
	}

	var r DynamoDBNamespaceService
	if err := mapstructure.Decode(svs[0].Credentials, &r); err != nil {
		return DynamoDBNamespaceService{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.AccessKeyID == "" || r.SecretAccessKey == "" || r.Region == "" || r.Prefix == "" {
		return DynamoDBNamespaceService{}, fmt.Errorf("parsed credentials are not valid")
	}

	return r, nil
}
