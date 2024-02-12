package credentials

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type Credentials struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	ARN             string `mapstructure:"arn"`
	URL             string `mapstructure:"queue_url"`
	Name            string `mapstructure:"queue_name"`
}

func Read() (Credentials, error) {
	app, err := cfenv.Current()
	if err != nil {
		return Credentials{}, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("sqs")
	if err != nil {
		return Credentials{}, fmt.Errorf("error reading Redis service details")
	}

	var r Credentials
	if err := mapstructure.Decode(svs[0].Credentials, &r); err != nil {
		return Credentials{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if err := r.validate(); err != nil {
		return Credentials{}, err
	}

	return r, nil
}

func (c Credentials) Config() (aws.Config, error) {
	return config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, ""))),
		config.WithRegion(c.Region),
	)
}

// validate checks every field in the binding that is expected to have a value
func (c Credentials) validate() error {
	var invalid []string
	v := reflect.ValueOf(c)
	t := v.Type()
	for i := range t.NumField() {
		if v.Field(i).String() == "" {
			invalid = append(invalid, t.Field(i).Name)
		}
	}

	switch len(invalid) {
	case 0:
		return nil
	default:
		return fmt.Errorf("parsed credentials are not valid, missing: %s", strings.Join(invalid, ", "))
	}
}
