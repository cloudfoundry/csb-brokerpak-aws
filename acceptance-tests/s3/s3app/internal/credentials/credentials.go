package credentials

import (
	"code.cloudfoundry.org/jsonry"
	"fmt"

	"os"
)

type S3Service struct {
	AccessKeyId     string `jsonry:"credentials.access_key_id"`
	AccessKeySecret string `jsonry:"credentials.secret_access_key"`
	Region  string    `jsonry:"credentials.region"`
	BucketDomainName  string    `jsonry:"credentials.bucket_domain_name"`
	BucketName  string    `jsonry:"credentials.bucket_name"`
	Arn  string    `jsonry:"credentials.arn"`
}

func Read() (*S3Service, error) {
	const variable = "VCAP_SERVICES"

	var services struct {
		S3Services []S3Service `jsonry:"csb-aws-s3-bucket"`
	}

	if err := jsonry.Unmarshal([]byte(os.Getenv(variable)), &services); err != nil {
		return nil, fmt.Errorf("failed to parse %q: %w", variable, err)
	}

	switch len(services.S3Services) {
	case 1: // ok
	case 0:
		return nil, fmt.Errorf("unable to find `csb-aws-s3-bucket` in %q", variable)
	default:
		return nil, fmt.Errorf("more than one entry for `csb-aws-s3-bucket` in %q", variable)
	}

	r := services.S3Services[0]
	if r.AccessKeyId == "" ||
		r.AccessKeySecret == "" ||
		r.Region == "" ||
		r.BucketName == "" ||
		r.BucketDomainName == "" ||
		r.Arn == "" {
		return nil, fmt.Errorf("parsed credentials are not valid: %s", os.Getenv(variable))
	}

	return &services.S3Services[0], nil
}
