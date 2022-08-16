package brokers

import (
	"fmt"
	"os"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"

	"github.com/onsi/ginkgo/v2"
)

func (b Broker) env() []apps.EnvVar {
	var result []apps.EnvVar

	for name, required := range map[string]bool{
		"AWS_ACCESS_KEY_ID":                      true,
		"AWS_SECRET_ACCESS_KEY":                  true,
		"GSB_BROKERPAK_BUILTIN_PATH":             false,
		"CH_CRED_HUB_URL":                        false,
		"CH_UAA_URL":                             false,
		"CH_UAA_CLIENT_NAME":                     false,
		"CH_UAA_CLIENT_SECRET":                   false,
		"CH_SKIP_SSL_VALIDATION":                 false,
		"GSB_COMPATIBILITY_ENABLE_BETA_SERVICES": false,
	} {
		val, ok := os.LookupEnv(name)
		switch {
		case ok:
			result = append(result, apps.EnvVar{Name: name, Value: val})
		case !ok && required:
			ginkgo.Fail(fmt.Sprintf("You must set the %s environment variable", name))
		}
	}

	result = append(result,
		apps.EnvVar{Name: "SECURITY_USER_NAME", Value: b.username},
		apps.EnvVar{Name: "SECURITY_USER_PASSWORD", Value: b.password},
		apps.EnvVar{Name: "DB_TLS", Value: "skip-verify"},
		apps.EnvVar{Name: "ENCRYPTION_ENABLED", Value: true},
		apps.EnvVar{Name: "ENCRYPTION_PASSWORDS", Value: b.secrets},
		apps.EnvVar{Name: "BROKERPAK_UPDATES_ENABLED", Value: true},
		apps.EnvVar{Name: "GSB_PROVISION_DEFAULTS", Value: fmt.Sprintf(`{"aws_vpc_id": %q}`, os.Getenv("AWS_PAS_VPC_ID"))},
	)

	return append(result, b.envExtras...)
}

func (b Broker) releaseEnv() []apps.EnvVar {
	return []apps.EnvVar{}
}

func (b Broker) latestEnv() []apps.EnvVar {
	return []apps.EnvVar{
		{Name: "GSB_COMPATIBILITY_ENABLE_BETA_SERVICES", Value: true},
		{Name: "GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS", Value: `[{"name":"default","id":"f64891b4-5021-4742-9871-dfe1a9051302","description":"Default S3 plan","display_name":"default"},{"name":"private","id":"8938b4c0-d67f-4c34-9f68-a66deef99b4e","description":"Private S3 bucket","display_name":"Private","acl":"private","boc_object_ownership":"ObjectWriter"}]`},
		{Name: "GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS", Value: `[{"name":"default","id":"de7dbcee-1c8d-11ed-9904-5f435c1e2316","description":"Default Postgres plan","display_name":"default", "instance_class": "db.m6i.large", "postgres_version": "11", "storage_gb": 10},{"name":"small","id":"ffc51616-228b-41bd-bed1-d601c18d58f5","description":"PostgreSQL 11, minimum 2 cores, minimum 4GB ram, 5GB storage","display_name":"small","storage_gb": 5, "cores": 2, "postgres_version": "11"}]`},
		{Name: "TERRAFORM_UPGRADES_ENABLED", Value: true},
	}
}
