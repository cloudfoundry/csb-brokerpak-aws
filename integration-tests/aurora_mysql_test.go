package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var customAuroraMySQLPlans = []map[string]any{
	customAuroraMySQLPlan,
}

var customAuroraMySQLPlan = map[string]any{
	"name":        "custom-sample",
	"id":          "10b2bd92-2a0b-11ed-b70f-c7c5cf3bb719",
	"description": "Default Aurora MySQL plan",
	"metadata": map[string]any{
		"displayName": "custom-sample",
	},
}

var _ = Describe("Aurora MySQL", Label("aurora-mysql"), func() {
	const serviceName = "csb-aws-aurora-mysql"

	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, serviceName)
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf("aws", "mysql", "aurora", "beta"))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal("custom-sample")}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should provision a plan", func() {
			instanceID, err := broker.Provision(serviceName, "custom-sample", nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				HaveKeyWithValue("instance_name", fmt.Sprintf("csb-auroramy-%s", instanceID)),
			)
		})
	})
})
