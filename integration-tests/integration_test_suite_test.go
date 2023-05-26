package integration_test

import (
	"encoding/json"
	"strings"
	"testing"

	"golang.org/x/exp/maps"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTests Suite")
}

const (
	awsSecretAccessKey = "aws-secret-access-key"
	awsAccessKeyID     = "aws-access-key-id"
	Name               = "Name"
	ID                 = "ID"
	fakeRegion         = "fake-region"
	documentationURL   = "https://docs.vmware.com/en/Cloud-Service-Broker-for-VMware-Tanzu/index.html"
)

var (
	mockTerraform testframework.TerraformMock
	broker        *testframework.TestInstance
)

var _ = BeforeSuite(func() {
	var err error
	mockTerraform, err = testframework.NewTerraformMock()
	Expect(err).NotTo(HaveOccurred())

	broker, err = testframework.BuildTestInstance(testframework.PathToBrokerPack(), mockTerraform, GinkgoWriter, "service-images")
	Expect(err).NotTo(HaveOccurred())

	Expect(broker.Start(GinkgoWriter, []string{
		"GSB_SERVICE_CSB_AWS_S3_BUCKET_PLANS=" + marshall(customS3Plans),
		"GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS=" + marshall(customPostgresPlans),
		"GSB_SERVICE_CSB_AWS_AURORA_POSTGRESQL_PLANS=" + marshall(customAuroraPostgresPlans),
		"GSB_SERVICE_CSB_AWS_AURORA_MYSQL_PLANS=" + marshall(customAuroraMySQLPlans),
		"GSB_SERVICE_CSB_AWS_MYSQL_PLANS=" + marshall(customMySQLPlans),
		"GSB_SERVICE_CSB_AWS_REDIS_PLANS=" + marshall(customRedisPlans),
		"GSB_SERVICE_CSB_AWS_MSSQL_PLANS=" + marshall(customMSSQLPlans),
		"AWS_ACCESS_KEY_ID=" + awsAccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + awsSecretAccessKey,
		"CSB_LISTENER_HOST=localhost",
		"GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=true",
		"GSB_PROVISION_DEFAULTS=" + marshall(map[string]string{"region": fakeRegion}),
	})).To(Succeed())
})

var _ = AfterSuite(func() {
	if broker != nil {
		Expect(broker.Cleanup()).To(Succeed())
	}
})

func marshall(element any) string {
	b, err := json.Marshal(element)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}

func stringOfLen(length int) string {
	return strings.Repeat("a", length)
}

func buildProperties(propertyOverrides ...map[string]any) map[string]any {
	result := map[string]any{}
	for _, override := range propertyOverrides {
		maps.Copy(result, override)
	}
	return result
}
