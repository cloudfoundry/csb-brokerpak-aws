package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	dynamoDBNamespaceServiceID                  = "07d06aeb-f87a-4e06-90ae-0b07a8c21a02"
	dynamoDBNamespaceServiceName                = "csb-aws-dynamodb-namespace"
	dynamoDBNamespaceServiceDescription         = "CSB Amazon DynamoDB Namespace"
	dynamoDBNamespaceServiceDisplayName         = "CSB Amazon DynamoDB Namespace"
	dynamoDBNamespaceServiceSupportURL          = "https://aws.amazon.com/dynamodb/"
	dynamoDBNamespaceServiceProviderDisplayName = "VMware"
)

var _ = Describe("DynamoDB Namespace", Label("DynamoDB Namespace"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish the service in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, dynamoDBNamespaceServiceName)
		Expect(service.ID).To(Equal(dynamoDBNamespaceServiceID))
		Expect(service.Description).To(Equal(dynamoDBNamespaceServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "dynamodb", "namespace"))
		Expect(service.Metadata.DisplayName).To(Equal(dynamoDBNamespaceServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(documentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(dynamoDBNamespaceServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(dynamoDBNamespaceServiceProviderDisplayName))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("default"),
					ID:   Equal("73b55e9a-4cdd-4d6f-81bd-c34d5c27a086"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should provision an instance", func() {
			instanceID, err := broker.Provision(dynamoDBNamespaceServiceName, "default", nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(SatisfyAll(
				HaveKeyWithValue("prefix", fmt.Sprintf("csb-%s-", instanceID)),
				HaveKeyWithValue("region", fakeRegion),
			))
		})
	})
})
