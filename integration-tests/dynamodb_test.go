package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	dynamoDBServiceID                  = "bf1db66a-1316-11eb-b959-e73b704ea230"
	dynamoDBServiceName                = "csb-aws-dynamodb"
	dynamoDBServiceDescription         = "Beta - CSB Amazon DynamoDB"
	dynamoDBServiceDisplayName         = "CSB Amazon DynamoDB (Beta)"
	dynamoDBServiceDocumentationURL    = "https://docs.vmware.com/en/Tanzu-Cloud-Service-Broker-for-AWS/1.2/csb-aws/GUID-reference-aws-dynamodb.html"
	dynamoDBServiceSupportURL          = "https://aws.amazon.com/dynamodb/"
	dynamoDBServiceProviderDisplayName = "VMware"
)

var _ = Describe("DynamoDB", Label("DynamoDB"), func() {
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

		service := testframework.FindService(catalog, dynamoDBServiceName)
		Expect(service.ID).To(Equal(dynamoDBServiceID))
		Expect(service.Description).To(Equal(dynamoDBServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "dynamodb", "beta"))
		Expect(service.Metadata.DisplayName).To(Equal(dynamoDBServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(dynamoDBServiceDocumentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(dynamoDBServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(dynamoDBServiceProviderDisplayName))
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
			_, err := broker.Provision(dynamoDBServiceName, "ondemand", attributes)

			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(dynamoDBServiceName, "ondemand", attributes)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should prevent updating region because it is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, dynamoDBServiceName, "ondemand", map[string]any{"region": "no-matter-what-region"})

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
