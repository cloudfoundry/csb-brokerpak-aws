package brokers

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/environment"
	"csbbrokerpakaws/acceptance-tests/helpers/testpath"
	"encoding/json"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	gsbProvisionDefaults = "GSB_PROVISION_DEFAULTS"
	gsbBrokerpakConfig   = "GSB_BROKERPAK_CONFIG"
)

func (b Broker) env() []apps.EnvVar {
	env := make(map[string]any)

	// Read these values from the environment
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
		"GSB_PROVISION_DEFAULTS":                 false,
	} {
		val, ok := os.LookupEnv(name)

		switch {
		case ok:
			env[name] = val
		case !ok && required:
			Fail(fmt.Sprintf("You must set the %s environment variable", name))
		}
	}

	// Default GSB_PROVISION_DEFAULTS if not set
	if _, ok := env[gsbProvisionDefaults]; !ok {
		env[gsbProvisionDefaults] = fmt.Sprintf(`{"aws_vpc_id": %q}`, os.Getenv("AWS_PAS_VPC_ID"))
	}

	env["SECURITY_USER_NAME"] = b.username
	env["SECURITY_USER_PASSWORD"] = b.password
	env["DB_TLS"] = "skip-verify"
	env["ENCRYPTION_ENABLED"] = true
	env["ENCRYPTION_PASSWORDS"] = b.secrets
	env["BROKERPAK_UPDATES_ENABLED"] = true
	env["GSB_COMPATIBILITY_ENABLE_BETA_SERVICES"] = true
	env["TERRAFORM_UPGRADES_ENABLED"] = true
	env["CSB_DISABLE_TF_UPGRADE_PROVIDER_RENAMES"] = false

	// Add extra environment variables, typically specific to a test or read from ".envrc"
	for _, e := range b.envExtras {
		env[e.Name] = e.Value
	}

	env[gsbBrokerpakConfig] = addGlobalLabels(env[gsbBrokerpakConfig], environment.ReadMetadata().Name)

	var result []apps.EnvVar
	for varName, value := range env {
		result = append(result, apps.EnvVar{Name: varName, Value: value})
	}

	return result
}

func (b Broker) latestEnv() []apps.EnvVar {
	return readEnvrcServices(testpath.BrokerpakFile(".envrc"))
}

// addGlobalLabels modifies the GSB_BROKERPAK_CONFIG environment variable if it exists,
// adding the JSON config required to make the broker add global labels. It labels resources
// with the name of the test environment that created it. This enables us to investigate any
// potential resource leaks
func addGlobalLabels(input any, environmentName string) map[string]any {
	result := make(map[string]any)
	switch i := input.(type) {
	case nil: // do nothing
	case string:
		if i != "" {
			Expect(json.Unmarshal([]byte(i), &result)).To(Succeed())
		}
	case map[string]any:
		result = i
	default:
		Fail(fmt.Sprintf("unexpected input type: %T", input))
	}

	labels := map[string]any{
		"key":   "origin",
		"value": environmentName,
	}

	arr, ok := result["global_labels"].([]map[string]any)
	switch ok {
	case true:
		result["global_labels"] = []map[string]any{labels}
	default:
		result["global_labels"] = append(arr, labels)
	}

	return result
}
