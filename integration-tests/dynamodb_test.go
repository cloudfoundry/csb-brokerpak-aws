package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("DynamoDB", Label("DynamoDB"), func() {
	const serviceName = "csb-aws-dynamodb"
	var attributes map[string]any

	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
		attributes = map[string]any{
			"table_name": "games",
			"hash_key":   "UserId",
			"range_key":  "GameTitle",
			"attributes": []any{},
			"global_secondary_indexes": []map[string]any{
				{
					"name":               "GameTitleIndex",
					"hash_key":           "GameTitle",
					"range_key":          "TopScore",
					"write_capacity":     10,
					"read_capacity":      10,
					"projection_type":    "INCLUDE",
					"non_key_attributes": []string{"UserId"},
				},
			},
		}
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish AWS dynamodb in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, serviceName)
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf("aws", "dynamodb", "beta"))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).To(HaveSuffix("(Beta)"))
		Expect(service.Metadata.LongDescription).To(HavePrefix("Beta -"))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal("ondemand")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("provisioned")}),
			),
		)
		Expect(service.Plans).To(
			HaveEach(
				MatchFields(IgnoreExtras, Fields{
					"Description": HavePrefix("Beta -"),
					"Metadata":    PointTo(MatchFields(IgnoreExtras, Fields{"DisplayName": HaveSuffix("(Beta)")})),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check region constraints", func() {
			attributes["region"] = "-Asia-northeast1"
			_, err := broker.Provision(serviceName, "ondemand", attributes)

			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(serviceName, "ondemand", attributes)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should prevent updating region because it is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, serviceName, "ondemand", map[string]any{"region": "no-matter-what-region"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))

			const initialProvisionInvocation = 1
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
		})
	})
})
