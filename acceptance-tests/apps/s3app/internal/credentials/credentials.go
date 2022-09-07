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

type S3ServiceLegacy struct {
	AccessKeyId     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"secret_access_key"`
	Region          string `mapstructure:"region"`
	BucketName      string `mapstructure:"bucket"`
}

func Read() (S3Service, error) {
	serviceTag, serviceCred, err := findService()
	if err != nil {
		return S3Service{}, err
	}
	switch serviceTag {
	case "s3":
		return ReadCSBS3(serviceCred)
	case "aws-s3":
		return ReadLegacyS3(serviceCred)
	}

	return S3Service{}, fmt.Errorf("unable to find credentials for S3")
}

func findService() (string, cfenv.Service, error) {
	app, err := cfenv.Current()
	if err != nil {
		return "", cfenv.Service{}, fmt.Errorf("error reading app env: %w", err)
	}

	for _, f := range []func() (string, []cfenv.Service, error){
		func() (string, []cfenv.Service, error) {
			serviceTag := "s3"
			srv, err := app.Services.WithTag(serviceTag)
			return serviceTag, srv, err

		},
		func() (string, []cfenv.Service, error) {
			serviceLabel := "aws-s3"
			srv, err := app.Services.WithLabel(serviceLabel)
			return serviceLabel, srv, err
		},
	} {
		serviceType, svs, err := f()
		if err == nil && len(svs) > 0 {
			return serviceType, svs[0], nil
		}
	}

	return "", cfenv.Service{}, fmt.Errorf("unable to find credentials for S3")
}

func ReadCSBS3(svs cfenv.Service) (S3Service, error) {
	var s S3Service
	if err := mapstructure.Decode(svs.Credentials, &s); err != nil {
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

func ReadLegacyS3(svs cfenv.Service) (S3Service, error) {
	var s S3ServiceLegacy
	if err := mapstructure.Decode(svs.Credentials, &s); err != nil {
		return S3Service{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if s.AccessKeyId == "" ||
		s.AccessKeySecret == "" ||
		s.Region == "" ||
		s.BucketName == "" {
		return S3Service{}, fmt.Errorf("parsed credentials are not valid")
	}

	return S3Service{
		AccessKeyId:     s.AccessKeyId,
		AccessKeySecret: s.AccessKeySecret,
		Region:          s.Region,
		BucketName:      s.BucketName,
	}, nil
}
