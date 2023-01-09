package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	redisServiceID                  = "e9c11b1b-0caa-45c9-b9b2-592939c9a5a6"
	redisServiceName                = "csb-aws-redis"
	redisServiceDescription         = "Beta - CSB Amazon ElastiCache for Redis - multinode with automatic failover"
	redisServiceDisplayName         = "CSB Amazon ElastiCache for Redis (Beta)"
	redisServiceDocumentationURL    = "https://docs.vmware.com/en/Tanzu-Cloud-Service-Broker-for-AWS/1.2/csb-aws/GUID-reference-aws-redis.html"
	redisServiceSupportURL          = "https://aws.amazon.com/redis/"
	redisServiceProviderDisplayName = "VMware"
)

var _ = Describe("Redis", Label("Redis"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish AWS redis in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, redisServiceName)
		Expect(service.ID).To(Equal(redisServiceID))
		Expect(service.Description).To(Equal(redisServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "redis", "beta"))
		Expect(service.Metadata.DisplayName).To(Equal(redisServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(redisServiceDocumentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(redisServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(redisServiceProviderDisplayName))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small"),
					ID:   Equal("ad963fcd-19f7-4b79-8e6d-645756e84f7a"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("df41095a-43e8-4be4-b4d6-ae2d8a35068d"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("da4dc49c-a64f-4d2a-8490-5e456cbb0577"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small-ha"),
					ID:   Equal("70544df7-0ac4-4580-ba51-c1fbdd6fdfd0"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium-ha"),
					ID:   Equal("a4235008-80f4-4053-924b-defcce17cb63"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large-ha"),
					ID:   Equal("f26cda6f-d4b4-473a-966c-32d238f723ef"),
				}),
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
			_, err := broker.Provision(redisServiceName, "small", map[string]any{"region": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(redisServiceName, "small", nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should prevent updating region because it is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, redisServiceName, "small", map[string]any{"region": "no-matter-what-region"})

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
