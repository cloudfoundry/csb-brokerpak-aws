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
		DescribeTable("should check property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(redisServiceName, "small", params)
				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"instance name minimum length is 6 characters",
				map[string]any{"instance_name": stringOfLen(5)},
				"instance_name: String length must be greater than or equal to 6",
			),
			Entry(
				"instance name maximum length is 40 characters",
				map[string]any{"instance_name": stringOfLen(99)},
				"instance_name: String length must be less than or equal to 40",
			),
			Entry(
				"instance name invalid characters",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
		)

		DescribeTable(
			"should prevent modifying `plan defined properties`",
			func(prop string, value any) {
				_, err := broker.Provision(redisServiceName, "small", map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"plan defined properties cannot be changed",
					),
				))

				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
			},
			Entry("cache_size", "cache_size", 9),
			Entry("node_count", "node_count", 5),
			Entry("redis_version", "redis_version", "4.0"),
		)

		DescribeTable(
			"should disallow `user_input` properties with the same name as some `computed_input` for clarity",
			func(prop string, value any) {
				_, err := broker.Provision(redisServiceName, "small", map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"additional properties are not allowed",
					),
				))

				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
			},
			Entry("labels", "labels", "a-valid-list-of-labels"),
		)

		DescribeTable("currently doesn't provide validations constraints for some properties",
			func(params map[string]any) {
				_, err := broker.Provision(redisServiceName, "small", params)
				Expect(err).NotTo(HaveOccurred())
			},
			Entry(
				"aws_vpc_id is never validated by the brokerpak logic",
				map[string]any{"aws_vpc_id": stringOfLen(99)},
			),
			Entry(
				"node_type is never validated by the brokerpak logic",
				map[string]any{"node_type": stringOfLen(99)},
			),
			Entry(
				"elasticache_subnet_group currently is never validated by the brokerpak logic",
				map[string]any{"elasticache_subnet_group": stringOfLen(99)},
			),
			Entry(
				"elasticache_vpc_security_group_ids is never validated by the brokerpak logic",
				map[string]any{"elasticache_vpc_security_group_ids": stringOfLen(99)},
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(redisServiceName, "small", map[string]any{})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKey("instance_name"),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
					HaveKeyWithValue("region", "us-west-2"),
					HaveKeyWithValue("cache_size", BeNumerically("==", 2)),
					HaveKeyWithValue("node_count", BeNumerically("==", 1)),
					HaveKeyWithValue("redis_version", "6.0"),
					HaveKeyWithValue("aws_vpc_id", BeEmpty()),
					HaveKeyWithValue("node_type", BeEmpty()),
					HaveKeyWithValue("elasticache_subnet_group", BeEmpty()),
					HaveKeyWithValue("elasticache_vpc_security_group_ids", BeEmpty()),
					HaveKeyWithValue("aws_access_key_id", "aws-access-key-id"),
					HaveKeyWithValue("aws_secret_access_key", "aws-secret-access-key"),
				))
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(redisServiceName, "small", map[string]any{
				"instance_name":                      "some-valid-instance-name",
				"region":                             "some-valid-region",
				"aws_vpc_id":                         "some-valid-aws-vpc-id",
				"node_type":                          "some-valid-node-type",
				"elasticache_subnet_group":           "some-valid-elasticache-subnet-group",
				"elasticache_vpc_security_group_ids": "some-valid-elasticache-vpc-security-group-ids",
				"aws_access_key_id":                  "some-valid-aws-access-key-id",
				"aws_secret_access_key":              "some-valid-aws-secret-access-key",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("node_count", BeNumerically("==", 1)),
					HaveKeyWithValue("redis_version", "6.0"),
					HaveKeyWithValue("cache_size", BeNumerically("==", 2)),
					HaveKeyWithValue("instance_name", "some-valid-instance-name"),
					HaveKeyWithValue("region", "some-valid-region"),
					HaveKeyWithValue("aws_vpc_id", "some-valid-aws-vpc-id"),
					HaveKeyWithValue("node_type", "some-valid-node-type"),
					HaveKeyWithValue("elasticache_subnet_group", "some-valid-elasticache-subnet-group"),
					HaveKeyWithValue("elasticache_vpc_security_group_ids", "some-valid-elasticache-vpc-security-group-ids"),
					HaveKeyWithValue("aws_access_key_id", "some-valid-aws-access-key-id"),
					HaveKeyWithValue("aws_secret_access_key", "some-valid-aws-secret-access-key"),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(redisServiceName, "small", nil)

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable(
			"preventing updates with `prohibit_update` as it can force resource replacement or re-creation",
			func(prop string, value any) {
				err := broker.Update(instanceID, redisServiceName, "small", map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("region", "region", "any-valid-value"),
			Entry("instance_name", "instance_name", "any-valid-instance-name"),
		)

		DescribeTable(
			"preventing updates for `plan defined properties` by design",
			func(prop string, value any) {
				err := broker.Update(instanceID, redisServiceName, "small", map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"plan defined properties cannot be changed",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("cache_size", "cache_size", 9),
			Entry("node_count", "node_count", 3),
			Entry("redis_version", "redis_version", "4.0"),
		)

		DescribeTable(
			"allowed updates",
			func(prop string, value any) {
				err := broker.Update(instanceID, redisServiceName, "small", map[string]any{prop: value})
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("aws_access_key_id", "aws_access_key_id", "any-valid-aws-access-key-id"),
			Entry("aws_secret_access_key", "aws_secret_access_key", "any-valid-aws-secret-access-key"),
			Entry("aws_vpc_id", "aws_vpc_id", "any-valid-aws-vpc-id"),
			Entry("node_type", "node_type", "any-valid-node-type"),
			Entry("elasticache_subnet_group", "elasticache_subnet_group", "any-valid-elasticache-subnet-group"),
			Entry("elasticache_vpc_security_group_ids", "elasticache_vpc_security_group_ids", "any-valid-elasticache-vpc-security-group-ids"),
		)
	})
})
