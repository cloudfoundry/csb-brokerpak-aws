package csbmajorengineversion_test

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-majorengineversion/csbmajorengineversion"
)

const (
	accessKeyID     = "AWS_ACCESS_KEY_ID"
	secretAccessKey = "AWS_SECRET_ACCESS_KEY"
	providerName    = "csbmajorengineversion"
)

var (
	tfStateDataResourceName = fmt.Sprintf("data.%s.major_version", csbmajorengineversion.DataResourceNameKey)
)

var _ = Describe("Provider", func() {
	var region = "us-west-2"
	DescribeTable("Major engine version can be obtained", func(engine, engineVersion, majorVersion, expectedErrorMessage string) {
		provider := initTestProvider()
		resource.Test(GinkgoT(), resource.TestCase{
			IsUnitTest:        true,
			ProviderFactories: getTestProviderFactories(provider),
			PreCheck: func() {
				failIfEnvEmpty(accessKeyID)
				failIfEnvEmpty(secretAccessKey)
			},
			Steps: []resource.TestStep{{
				Config: testGetConfiguration(
					os.Getenv(accessKeyID),
					os.Getenv(secretAccessKey),
					region,
					engine,
					engineVersion,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfStateDataResourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(tfStateDataResourceName, "id", "version"),
					resource.TestCheckResourceAttr(tfStateDataResourceName, "major_version", majorVersion),
				),
			}},
			ErrorCheck: func(err error) error {
				if err != nil {
					if expectedErrorMessage == "" {
						return err
					}

					if strings.Contains(err.Error(), expectedErrorMessage) {
						return nil
					}

					return fmt.Errorf("Terraform provider %s error check fails: %w", providerName, err)
				}
				return nil
			},
		})
	},
		Entry("postgres 14", "postgres", "14.2", "14", ""), // --include-all option. 14.2 is not in the list
		Entry("postgres 15", "postgres", "15.3", "15", ""),
		Entry("aurora-mysql 8", "aurora-mysql", "8.0", "8.0", ""),
		Entry("aurora-mysql 8.0.mysql_aurora.3.03.1", "aurora-mysql", "8.0.mysql_aurora.3.03.1", "8.0", ""),
		Entry("aurora-postgresql 14", "aurora-postgresql", "14", "14", ""),
		Entry("aurora-postgresql 14.3", "aurora-postgresql", "14.3", "14", ""),
		Entry("mysql 5.7", "mysql", "5.7", "5.7", ""),
		Entry("mysql 5.7.42", "mysql", "5.7.42", "5.7", ""),
		Entry("mysql 8.0", "mysql", "8.0", "8.0", ""),
		Entry("mysql 8.0.32", "mysql", "8.0.32", "8.0", ""),
		Entry("no engine", "", "", "8.0.32", `Error: expected "engine" to not be an empty string`),
		Entry("no engine version", "mysql", "", "", `Error: expected "engine_version" to not be an empty string`),
		Entry(
			"aurora-postgresql 8.0.postgresql_aurora.3.02.0",
			"aurora-postgresql",
			"8.0.postgresql_aurora.3.02.0",
			"xx",
			"Error: invalid parameter combination. API does not return any db engine version - engine aurora-postgresql - engine version 8.0.postgresql_aurora.3.02.0",
		),
	)

})

func initTestProvider() *schema.Provider {
	testAccProvider := &schema.Provider{
		Schema: csbmajorengineversion.ProviderSchema(),
		DataSourcesMap: map[string]*schema.Resource{
			csbmajorengineversion.DataResourceNameKey: csbmajorengineversion.DataSourceMajorEngineVersion(),
		},
		ConfigureContextFunc: csbmajorengineversion.ProviderConfigureContext,
	}
	err := testAccProvider.InternalValidate()
	Expect(err).NotTo(HaveOccurred())

	return testAccProvider
}

func getTestProviderFactories(provider *schema.Provider) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		providerName: func() (*schema.Provider, error) {
			if provider == nil {
				return provider, errors.New("provider cannot be nil")
			}

			return provider, nil
		},
	}
}

func testGetConfiguration(accessKeyID, secretAccessKey, region, engine, engineVersion string) string {
	return fmt.Sprintf(`
provider "csbmajorengineversion" {
  engine 			= %[1]q
  access_key_id     = %[2]q
  secret_access_key = %[3]q
  region            = %[4]q
}

data "csbmajorengineversion" "major_version" {
  engine_version     = %[5]q
}
`, engine, accessKeyID, secretAccessKey, region, engineVersion)
}

func failIfEnvEmpty(name string) {
	value := os.Getenv(name)
	Expect(value).NotTo(BeEmpty(), "environment variable %s must be set.", name)
}
