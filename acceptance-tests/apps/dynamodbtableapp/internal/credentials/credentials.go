package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type DynamoDBService struct {
	AccessKeyId     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	TableName       string `mapstructure:"dynamodb_table_name"`
}

func Read() (DynamoDBService, error) {
	app, err := cfenv.Current()
	if err != nil {
		return DynamoDBService{}, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("dynamodb")
	if err != nil {
		return DynamoDBService{}, fmt.Errorf("error reading DynamoDB service details")
	}

	var r DynamoDBService
	if err := mapstructure.Decode(svs[0].Credentials, &r); err != nil {
		return DynamoDBService{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.AccessKeyId == "" || r.AccessKeySecret == "" || r.Region == "" || r.TableName == "" {
		return DynamoDBService{}, fmt.Errorf("parsed credentials are not valid")
	}

	return r, nil
}
