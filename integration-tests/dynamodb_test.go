package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

		It("should allow updating region because it is not flagged as `prohibit_update` and not specified in the plan", func() {
			err := broker.Update(instanceID, serviceName, "ondemand", map[string]any{"region": "asia-southeast1"})

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
