package brokers

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"fmt"
	"os"

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
		apps.EnvVar{Name: "GSB_COMPATIBILITY_ENABLE_BETA_SERVICES", Value: true},
		apps.EnvVar{Name: "GSB_PROVISION_DEFAULTS", Value: fmt.Sprintf(`{"aws_vpc_id": %q}`, os.Getenv("AWS_PAS_VPC_ID"))},
	)

	return append(result, b.envExtras...)
}
