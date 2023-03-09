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
	redisCustomPlanName             = "custom-sample"
	redisCustomPlanID               = "c7f64994-a1d9-4e1f-9491-9d8e56bbf146"
)

var customRedisPlans = []map[string]any{
	customRedisPlan,
}

var customRedisPlan = map[string]any{
	"name":        redisCustomPlanName,
	"id":          redisCustomPlanID,
	"description": "Beta - Default Redis plan",
	"cache_size":  2,
	"node_count":  2,
	"metadata": map[string]any{
		"displayName": "custom-sample (Beta)",
	},
}

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
					Name: Equal(redisCustomPlanName),
					ID:   Equal("c7f64994-a1d9-4e1f-9491-9d8e56bbf146"),
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
				_, err := broker.Provision(redisServiceName, redisCustomPlanName, params)
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
				map[string]any{"instance_name": stringOfLen(41)},
				"instance_name: String length must be less than or equal to 40",
			),
			Entry(
				"instance name invalid characters",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
		)

		It("should prevent modifying `plan defined properties`", func() {
			_, err := broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{"cache_size": 9})

			Expect(err).To(MatchError(
				ContainSubstring(
					"plan defined properties cannot be changed",
				),
			))

			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})

		DescribeTable(
			"should disallow `user_input` properties with the same name as some `computed_input` for clarity",
			func(prop string, value any) {
				_, err := broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"additional properties are not allowed",
					),
				))

				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
			},
			Entry("labels", "labels", "a-valid-list-of-labels"),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{"redis_version": "6.x"})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKey("instance_name"),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
					HaveKeyWithValue("region", "us-west-2"),
					HaveKeyWithValue("cache_size", BeNumerically("==", 2)),
					HaveKeyWithValue("node_count", BeNumerically("==", 2)),
					HaveKeyWithValue("redis_version", "6.x"),
					HaveKeyWithValue("aws_vpc_id", BeEmpty()),
					HaveKeyWithValue("node_type", BeEmpty()),
					HaveKeyWithValue("elasticache_subnet_group", BeEmpty()),
					HaveKeyWithValue("elasticache_vpc_security_group_ids", BeEmpty()),
					HaveKeyWithValue("aws_access_key_id", "aws-access-key-id"),
					HaveKeyWithValue("aws_secret_access_key", "aws-secret-access-key"),
					HaveKeyWithValue("at_rest_encryption_enabled", BeTrue()),
					HaveKeyWithValue("kms_key_id", ""),
				))
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{
				"redis_version":                      "6.x",
				"instance_name":                      "some-valid-instance-name",
				"region":                             "some-valid-region",
				"aws_vpc_id":                         "some-valid-aws-vpc-id",
				"node_type":                          "some-valid-node-type",
				"elasticache_subnet_group":           "some-valid-elasticache-subnet-group",
				"elasticache_vpc_security_group_ids": "some-valid-elasticache-vpc-security-group-ids",
				"aws_access_key_id":                  "some-valid-aws-access-key-id",
				"aws_secret_access_key":              "some-valid-aws-secret-access-key",
				"at_rest_encryption_enabled":         false,
				"kms_key_id":                         "fake-encryption-at-rest-key",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("node_count", BeNumerically("==", 2)),
					HaveKeyWithValue("redis_version", "6.x"),
					HaveKeyWithValue("cache_size", BeNumerically("==", 2)),
					HaveKeyWithValue("instance_name", "some-valid-instance-name"),
					HaveKeyWithValue("region", "some-valid-region"),
					HaveKeyWithValue("aws_vpc_id", "some-valid-aws-vpc-id"),
					HaveKeyWithValue("node_type", "some-valid-node-type"),
					HaveKeyWithValue("elasticache_subnet_group", "some-valid-elasticache-subnet-group"),
					HaveKeyWithValue("elasticache_vpc_security_group_ids", "some-valid-elasticache-vpc-security-group-ids"),
					HaveKeyWithValue("aws_access_key_id", "some-valid-aws-access-key-id"),
					HaveKeyWithValue("aws_secret_access_key", "some-valid-aws-secret-access-key"),
					HaveKeyWithValue("at_rest_encryption_enabled", BeFalse()),
					HaveKeyWithValue("kms_key_id", "fake-encryption-at-rest-key"),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{"redis_version": "6.x"})

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable(
			"preventing updates with `prohibit_update` as it can force resource replacement or re-creation",
			func(prop string, value any) {
				err := broker.Update(instanceID, redisServiceName, redisCustomPlanName, map[string]any{prop: value})

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
			Entry("at_rest_encryption_enabled", "at_rest_encryption_enabled", false),
			Entry("kms_key_id", "kms_key_id", "fake-encryption-at-rest-key"),
		)

		It("preventing updates for `plan defined properties` by design", func() {
			err := broker.Update(instanceID, redisServiceName, redisCustomPlanName, map[string]any{"cache_size": 9})

			Expect(err).To(MatchError(
				ContainSubstring(
					"plan defined properties cannot be changed",
				),
			))

			const initialProvisionInvocation = 1
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
		})

		DescribeTable(
			"allowed updates",
			func(prop string, value any) {
				err := broker.Update(instanceID, redisServiceName, redisCustomPlanName, map[string]any{prop: value})
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("aws_access_key_id", "aws_access_key_id", "any-valid-aws-access-key-id"),
			Entry("aws_secret_access_key", "aws_secret_access_key", "any-valid-aws-secret-access-key"),
			Entry("aws_vpc_id", "aws_vpc_id", "any-valid-aws-vpc-id"),
			Entry("node_type", "node_type", "any-valid-node-type"),
			Entry("elasticache_subnet_group", "elasticache_subnet_group", "any-valid-elasticache-subnet-group"),
			Entry("elasticache_vpc_security_group_ids", "elasticache_vpc_security_group_ids", "any-valid-elasticache-vpc-security-group-ids"),
			Entry("redis_version", "redis_version", "7.x"),
		)
	})
})
