package integration_test

import (
	"encoding/json"
	"fmt"
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
		`GSB_BROKERPAK_CONFIG={"global_labels":[{"key":  "key1", "value":  "value1"},{"key":  "key2", "value":  "value2"}]}`,
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

func deleteProperty(key string, properties map[string]any) map[string]any {
	delete(properties, key)
	return properties
}

func nthTerraformInvocationVars(p testframework.TerraformMock, n int) (map[string]any, error) {
	// Notice that n is expected to be zero-based
	// Copied from https://github.com/cloudfoundry/cloud-service-broker/blob/0a9138aa6cbda08e9bf1f8c32feef16a9a610956/brokerpaktestframework/terraform_mock.go#L101-L115
	invocations, err := p.ApplyInvocations()
	if err != nil {
		return nil, err
	}
	if len(invocations) < n+1 || n < 0 {
		return nil, fmt.Errorf("unexpected invocation index. max_index: %d requested_index: %d", len(invocations)-1, n)
	}

	vars, err := invocations[n].TFVars()
	if err != nil {
		return nil, err
	}
	return vars, nil
}
