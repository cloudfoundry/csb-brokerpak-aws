package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type S3Service struct {
	AccessKeyId      string `mapstructure:"access_key_id"`
	AccessKeySecret  string `mapstructure:"secret_access_key"`
	Region           string `mapstructure:"region"`
	BucketDomainName string `mapstructure:"bucket_domain_name"`
	BucketName       string `mapstructure:"bucket_name"`
	Arn              string `mapstructure:"arn"`
}

func Read() (S3Service, error) {
	app, err := cfenv.Current()
	if err != nil {
		return S3Service{}, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("s3")
	if err != nil {
		return S3Service{}, fmt.Errorf("error reading S3 service details")
	}

	var s S3Service

	if err := mapstructure.Decode(svs[0].Credentials, &s); err != nil {
		return S3Service{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if s.AccessKeyId == "" ||
		s.AccessKeySecret == "" ||
		s.Region == "" ||
		s.BucketName == "" ||
		s.BucketDomainName == "" ||
		s.Arn == "" {
		return S3Service{}, fmt.Errorf("parsed credentials are not valid")
	}

	return s, nil
}
