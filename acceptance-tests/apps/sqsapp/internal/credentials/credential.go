package credentials

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type Credential struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	ARN             string `mapstructure:"arn"`
	URL             string `mapstructure:"queue_url"`
	Name            string `mapstructure:"queue_name"`
}

func (c Credential) Config() (aws.Config, error) {
	return config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, ""))),
		config.WithRegion(c.Region),
	)
}

// validate checks every field in the binding that is expected to have a value
func (c Credential) validate() error {
	var invalid []string
	v := reflect.ValueOf(c)
	t := v.Type()
	for i := range t.NumField() {
		if v.Field(i).String() == "" {
			invalid = append(invalid, t.Field(i).Name)
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("parsed credentials are not valid, missing: %s", strings.Join(invalid, ", "))
	}

	return nil
}
