package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	dynamoDBTableServiceID                  = "bf1db66a-1316-11eb-b959-e73b704ea230"
	dynamoDBTableServiceName                = "csb-aws-dynamodb-table"
	dynamoDBTableServiceDescription         = "Beta - CSB Amazon DynamoDB Table"
	dynamoDBTableServiceDisplayName         = "CSB Amazon DynamoDB Table (Beta)"
	dynamoDBTableServiceDocumentationURL    = "https://docs.vmware.com/en/Tanzu-Cloud-Service-Broker-for-AWS/1.2/csb-aws/GUID-reference-aws-dynamodb.html"
	dynamoDBTableServiceSupportURL          = "https://aws.amazon.com/dynamodb/"
	dynamoDBTableServiceProviderDisplayName = "VMware"
)

var _ = Describe("DynamoDB Table", Label("DynamoDB Table"), func() {
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

		service := testframework.FindService(catalog, dynamoDBTableServiceName)
		Expect(service.ID).To(Equal(dynamoDBTableServiceID))
		Expect(service.Description).To(Equal(dynamoDBTableServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "dynamodb", "dynamodb-table", "beta"))
		Expect(service.Metadata.DisplayName).To(Equal(dynamoDBTableServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(dynamoDBTableServiceDocumentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(dynamoDBTableServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(dynamoDBTableServiceProviderDisplayName))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name:          Equal("ondemand"),
					ID:            Equal("52b109ee-1318-11eb-851b-dbe6aa707e6b"),
					"Description": HavePrefix("Beta -"),
					"Metadata":    PointTo(MatchFields(IgnoreExtras, Fields{"DisplayName": HaveSuffix("(Beta)")})),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name:          Equal("provisioned"),
					ID:            Equal("591808b4-1318-11eb-b932-cbf259c3124c"),
					"Description": HavePrefix("Beta -"),
					"Metadata":    PointTo(MatchFields(IgnoreExtras, Fields{"DisplayName": HaveSuffix("(Beta)")})),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check region constraints", func() {
			attributes["region"] = "-Asia-northeast1"
			_, err := broker.Provision(dynamoDBTableServiceName, "ondemand", attributes)

			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(dynamoDBTableServiceName, "ondemand", attributes)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should prevent updating region because it is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, dynamoDBTableServiceName, "ondemand", map[string]any{"region": "no-matter-what-region"})

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
