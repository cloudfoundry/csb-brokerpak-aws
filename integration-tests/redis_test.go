package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	redisServiceID                        = "e9c11b1b-0caa-45c9-b9b2-592939c9a5a6"
	redisServiceName                      = "csb-aws-redis"
	redisServiceDescription               = "CSB Amazon ElastiCache for Redis"
	redisServiceDisplayName               = "CSB Amazon ElastiCache for Redis"
	redisServiceSupportURL                = "https://aws.amazon.com/redis/"
	redisServiceProviderDisplayName       = "VMware"
	redisCustomPlanName                   = "custom-sample"
	redisCustomPlanID                     = "c7f64994-a1d9-4e1f-9491-9d8e56bbf146"
	deprecatedCacheSizePlanName           = "deprecated-cachesize-sample"
	deprecatedCacheSizePlanID             = "eeae19c8-00c1-442d-b423-a377684b70df"
	redisPlanWithFlexibleNodeTypePlanName = "flexible-nodetype-sample"
	redisPlanWithFlexibleNodeTypePlanID   = "2deb6c13-7ea1-4bad-a519-0ac9600e9a29"
)

var customRedisPlans = []map[string]any{
	customRedisPlan,
	redisPlanWithFlexibleNodeType,
	deprecatedCacheSizePlan,
}

var customRedisPlan = map[string]any{
	"name":        redisCustomPlanName,
	"id":          redisCustomPlanID,
	"description": "Default Redis plan",
	"node_type":   "cache.t3.medium",
	"node_count":  2,
	"metadata": map[string]any{
		"displayName": "custom-sample",
	},
}

var redisPlanWithFlexibleNodeType = map[string]any{
	"name":          redisPlanWithFlexibleNodeTypePlanName,
	"id":            redisPlanWithFlexibleNodeTypePlanID,
	"description":   "An example of a Redis plan for which node_type can be specified at provision time.",
	"redis_version": "6.x",
	"node_count":    2,
	"metadata": map[string]any{
		"displayName": "flexible-nodetype-sample",
	},
}

var deprecatedCacheSizePlan = map[string]any{
	"name":          deprecatedCacheSizePlanName,
	"id":            deprecatedCacheSizePlanID,
	"description":   "Redis plan with deprecated cache_size",
	"cache_size":    2,
	"redis_version": "6.x",
	"node_count":    2,
	"metadata": map[string]any{
		"displayName": "deprecated-cachesize-sample",
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
		Expect(service.Tags).To(ConsistOf("aws", "redis"))
		Expect(service.Metadata.DisplayName).To(Equal(redisServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(documentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(redisServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(redisServiceProviderDisplayName))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal(redisCustomPlanName),
					ID:   Equal("c7f64994-a1d9-4e1f-9491-9d8e56bbf146"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal(deprecatedCacheSizePlanName),
					ID:   Equal(deprecatedCacheSizePlanID),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal(redisPlanWithFlexibleNodeTypePlanName),
					ID:   Equal(redisPlanWithFlexibleNodeTypePlanID),
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
			Entry(
				"maintenance_day invalid day",
				map[string]any{"maintenance_day": "San"},
				`maintenance_day must be one of the following: null, \"Fri\", \"Mon\", \"Sat\", \"Sun\", \"Thu\", \"Tue\", \"Wed\"`,
			),
			Entry(
				"maintenance_start_hour invalid hour",
				map[string]any{"maintenance_start_hour": "31"},
				`maintenance_start_hour must be one of the following: \"00\", \"01\", \"02\", \"03\", \"04\", \"05\", \"06\", \"07\", \"08\", \"09\", \"10\", \"11\", \"12\", \"13\", \"14\", \"15\", \"16\", \"17\", \"18\", \"19\", \"20\", \"21\", \"22\", \"23\", null"`,
			),
			Entry(
				"maintenance_start_min invalid minute",
				map[string]any{"maintenance_start_min": "12"},
				`maintenance_start_min must be one of the following: \"00\", \"15\", \"30\", \"45\", null`,
			),
			Entry(
				"maintenance_end_hour invalid hour",
				map[string]any{"maintenance_end_hour": "31"},
				`maintenance_end_hour must be one of the following: \"00\", \"01\", \"02\", \"03\", \"04\", \"05\", \"06\", \"07\", \"08\", \"09\", \"10\", \"11\", \"12\", \"13\", \"14\", \"15\", \"16\", \"17\", \"18\", \"19\", \"20\", \"21\", \"22\", \"23\", null"`,
			),
			Entry(
				"maintenance_end_min invalid minute",
				map[string]any{"maintenance_end_min": "12"},
				`maintenance_end_min must be one of the following: \"00\", \"15\", \"30\", \"45\", null`,
			),
			Entry(
				"backup_start_hour invalid hour",
				map[string]any{"backup_start_hour": "31"},
				`backup_start_hour must be one of the following: \"00\", \"01\", \"02\", \"03\", \"04\", \"05\", \"06\", \"07\", \"08\", \"09\", \"10\", \"11\", \"12\", \"13\", \"14\", \"15\", \"16\", \"17\", \"18\", \"19\", \"20\", \"21\", \"22\", \"23\", null"`,
			),
			Entry(
				"backup_start_min invalid minute",
				map[string]any{"backup_start_min": "12"},
				`backup_start_min must be one of the following: \"00\", \"15\", \"30\", \"45\", null`,
			),
			Entry(
				"backup_end_hour invalid hour",
				map[string]any{"backup_end_hour": "31"},
				`backup_end_hour must be one of the following: \"00\", \"01\", \"02\", \"03\", \"04\", \"05\", \"06\", \"07\", \"08\", \"09\", \"10\", \"11\", \"12\", \"13\", \"14\", \"15\", \"16\", \"17\", \"18\", \"19\", \"20\", \"21\", \"22\", \"23\", null"`,
			),
			Entry(
				"backup_end_min invalid minute",
				map[string]any{"backup_end_min": "12"},
				`backup_end_min must be one of the following: \"00\", \"15\", \"30\", \"45\", null`,
			),
		)

		It("should prevent modifying `plan defined properties`", func() {
			_, err := broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{"node_count": 3})

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

		It("should prevent passing the deprecated property `cache_size` as user input", func() {
			_, err := broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{"cache_size": 2})
			Expect(err).To(MatchError(ContainSubstring("additional properties are not allowed: cache_size")))
		})

		It("should keep working as before for existing customers relying on `cache_size` property", func() {
			_, err := broker.Provision(redisServiceName, deprecatedCacheSizePlanName, map[string]any{})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("cache_size", BeNumerically("==", 2)),
				))
		})

		It("should keep working as before for existing customers relying on `cache_size` property and per-instance `node_type`", func() {
			_, err := broker.Provision(redisServiceName, deprecatedCacheSizePlanName, map[string]any{"node_type": "cache.t2.micro"})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("cache_size", BeNumerically("==", 2)),
					HaveKeyWithValue("node_type", "cache.t2.micro"),
				))
		})

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{"redis_version": "6.x"})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKey("instance_name"),
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{
						"pcf-instance-id": Equal(instanceID),
						"key1":            Equal("value1"),
						"key2":            Equal("value2"),
					})),
					HaveKeyWithValue("region", fakeRegion),
					HaveKeyWithValue("cache_size", BeNil()),
					HaveKeyWithValue("node_count", BeNumerically("==", 2)),
					HaveKeyWithValue("redis_version", "6.x"),
					HaveKeyWithValue("aws_vpc_id", BeEmpty()),
					HaveKeyWithValue("node_type", "cache.t3.medium"),
					HaveKeyWithValue("elasticache_subnet_group", BeEmpty()),
					HaveKeyWithValue("elasticache_vpc_security_group_ids", BeEmpty()),
					HaveKeyWithValue("aws_access_key_id", "aws-access-key-id"),
					HaveKeyWithValue("aws_secret_access_key", "aws-secret-access-key"),
					HaveKeyWithValue("at_rest_encryption_enabled", BeTrue()),
					HaveKeyWithValue("kms_key_id", ""),
					HaveKeyWithValue("maintenance_day", BeNil()),
					HaveKeyWithValue("maintenance_start_hour", BeNil()),
					HaveKeyWithValue("maintenance_start_min", BeNil()),
					HaveKeyWithValue("maintenance_end_hour", BeNil()),
					HaveKeyWithValue("maintenance_end_min", BeNil()),
					HaveKeyWithValue("data_tiering_enabled", BeFalse()),
					HaveKeyWithValue("automatic_failover_enabled", BeTrue()),
					HaveKeyWithValue("multi_az_enabled", BeTrue()),
					HaveKeyWithValue("backup_retention_limit", BeNumerically("==", 1)),
					HaveKeyWithValue("final_backup_identifier", BeNil()),
					HaveKeyWithValue("backup_name", ""),
					HaveKeyWithValue("backup_start_hour", BeNil()),
					HaveKeyWithValue("backup_start_min", BeNil()),
					HaveKeyWithValue("backup_end_hour", BeNil()),
					HaveKeyWithValue("backup_end_min", BeNil()),
					HaveKeyWithValue("parameter_group_name", ""),
					HaveKeyWithValue("preferred_azs", BeEmpty()),
					HaveKeyWithValue("logs_slow_log_enabled", BeFalse()),
					HaveKeyWithValue("logs_slow_log_loggroup_retention_in_days", BeNumerically("==", 0)),
					HaveKeyWithValue("logs_slow_log_loggroup_kms_key_id", BeEmpty()),
					HaveKeyWithValue("logs_engine_log_enabled", BeFalse()),
					HaveKeyWithValue("logs_engine_log_loggroup_retention_in_days", BeNumerically("==", 0)),
					HaveKeyWithValue("logs_engine_log_loggroup_kms_key_id", BeEmpty()),
					HaveKeyWithValue("auto_minor_version_upgrade", BeFalse()),
				))
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(redisServiceName, redisCustomPlanName, map[string]any{
				"redis_version":                              "6.x",
				"instance_name":                              "some-valid-instance-name",
				"region":                                     "some-valid-region",
				"aws_vpc_id":                                 "some-valid-aws-vpc-id",
				"elasticache_subnet_group":                   "some-valid-elasticache-subnet-group",
				"elasticache_vpc_security_group_ids":         "some-valid-elasticache-vpc-security-group-ids",
				"aws_access_key_id":                          "some-valid-aws-access-key-id",
				"aws_secret_access_key":                      "some-valid-aws-secret-access-key",
				"at_rest_encryption_enabled":                 false,
				"kms_key_id":                                 "fake-encryption-at-rest-key",
				"maintenance_day":                            "Mon",
				"maintenance_start_hour":                     "03",
				"maintenance_start_min":                      "45",
				"maintenance_end_hour":                       "10",
				"maintenance_end_min":                        "15",
				"automatic_failover_enabled":                 false,
				"multi_az_enabled":                           false,
				"backup_retention_limit":                     32,
				"final_backup_identifier":                    "tortoise",
				"backup_name":                                "turtle",
				"backup_start_hour":                          "04",
				"backup_start_min":                           "15",
				"backup_end_hour":                            "11",
				"backup_end_min":                             "30",
				"parameter_group_name":                       "fake-param-group-name",
				"preferred_azs":                              []string{"fake-az1", "fake-az2"},
				"logs_slow_log_enabled":                      true,
				"logs_slow_log_loggroup_retention_in_days":   1,
				"logs_slow_log_loggroup_kms_key_id":          "slow-log-key",
				"logs_engine_log_enabled":                    true,
				"logs_engine_log_loggroup_retention_in_days": 2,
				"logs_engine_log_loggroup_kms_key_id":        "engine-log-key",
				"auto_minor_version_upgrade":                 true,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("node_count", BeNumerically("==", 2)),
					HaveKeyWithValue("redis_version", "6.x"),
					HaveKeyWithValue("cache_size", BeNil()),
					HaveKeyWithValue("instance_name", "some-valid-instance-name"),
					HaveKeyWithValue("region", "some-valid-region"),
					HaveKeyWithValue("aws_vpc_id", "some-valid-aws-vpc-id"),
					HaveKeyWithValue("node_type", "cache.t3.medium"),
					HaveKeyWithValue("elasticache_subnet_group", "some-valid-elasticache-subnet-group"),
					HaveKeyWithValue("elasticache_vpc_security_group_ids", "some-valid-elasticache-vpc-security-group-ids"),
					HaveKeyWithValue("aws_access_key_id", "some-valid-aws-access-key-id"),
					HaveKeyWithValue("aws_secret_access_key", "some-valid-aws-secret-access-key"),
					HaveKeyWithValue("at_rest_encryption_enabled", BeFalse()),
					HaveKeyWithValue("kms_key_id", "fake-encryption-at-rest-key"),
					HaveKeyWithValue("maintenance_day", "Mon"),
					HaveKeyWithValue("maintenance_start_hour", "03"),
					HaveKeyWithValue("maintenance_start_min", "45"),
					HaveKeyWithValue("maintenance_end_hour", "10"),
					HaveKeyWithValue("maintenance_end_min", "15"),
					HaveKeyWithValue("automatic_failover_enabled", BeFalse()),
					HaveKeyWithValue("multi_az_enabled", BeFalse()),
					HaveKeyWithValue("backup_name", "turtle"),
					HaveKeyWithValue("backup_start_hour", "04"),
					HaveKeyWithValue("backup_start_min", "15"),
					HaveKeyWithValue("backup_end_hour", "11"),
					HaveKeyWithValue("backup_end_min", "30"),
					HaveKeyWithValue("parameter_group_name", "fake-param-group-name"),
					HaveKeyWithValue("preferred_azs", ConsistOf("fake-az1", "fake-az2")),
					HaveKeyWithValue("logs_slow_log_enabled", BeTrue()),
					HaveKeyWithValue("logs_slow_log_loggroup_retention_in_days", BeNumerically("==", 1)),
					HaveKeyWithValue("logs_slow_log_loggroup_kms_key_id", "slow-log-key"),
					HaveKeyWithValue("logs_engine_log_enabled", BeTrue()),
					HaveKeyWithValue("logs_engine_log_loggroup_retention_in_days", BeNumerically("==", 2)),
					HaveKeyWithValue("logs_engine_log_loggroup_kms_key_id", "engine-log-key"),
					HaveKeyWithValue("auto_minor_version_upgrade", BeTrue()),
				),
			)
		})

		It("`data_tiering_enabled` must be set to true when using `r6gd` nodes when provisioning", func() {
			_, err := broker.Provision(redisServiceName, redisPlanWithFlexibleNodeTypePlanName, map[string]any{
				"node_type":            "cache.r6gd.xlarge",
				"data_tiering_enabled": true,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("node_type", "cache.r6gd.xlarge"),
					HaveKeyWithValue("data_tiering_enabled", true),
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
			Entry("data_tiering_enabled", "data_tiering_enabled", true),
			Entry("backup_name", "backup_name", "turtle"),
			Entry("preferred_azs", "preferred_azs", []string{"fake-az1", "fake-az2"}),
			Entry("aws_vpc_id", "aws_vpc_id", "any-valid-aws-vpc-id"),
			Entry("elasticache_subnet_group", "elasticache_subnet_group", "any-valid-elasticache-subnet-group"),
			Entry("elasticache_vpc_security_group_ids", "elasticache_vpc_security_group_ids", "any-valid-elasticache-vpc-security-group-ids"),
		)

		It("preventing updates for `plan defined properties` by design", func() {
			err := broker.Update(instanceID, redisServiceName, redisCustomPlanName, map[string]any{"node_type": "cache.t2.micro"})

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
			Entry("redis_version", "redis_version", "7.x"),
			Entry("automatic_failover_enabled", "automatic_failover_enabled", false),
			Entry("backup_retention_limit", "backup_retention_limit", 12),
			Entry("final_backup_identifier", "final_backup_identifier", "tank"),
			Entry("maintenance_day", "maintenance_day", "Wed"),
			Entry("maintenance_start_hour", "maintenance_start_hour", "05"),
			Entry("maintenance_start_min", "maintenance_start_min", "30"),
			Entry("maintenance_end_hour", "maintenance_end_hour", "05"),
			Entry("maintenance_end_min", "maintenance_end_min", "30"),
			Entry("backup_start_hour", "backup_start_hour", "05"),
			Entry("backup_start_min", "backup_start_min", "30"),
			Entry("backup_end_hour", "backup_end_hour", "05"),
			Entry("backup_end_min", "backup_end_min", "30"),
			Entry("parameter_group_name", "parameter_group_name", "fake-param-group-name"),
			Entry("logs_slow_log_enabled", "logs_slow_log_enabled", true),
			Entry("logs_slow_log_loggroup_retention_in_days", "logs_slow_log_loggroup_retention_in_days", 4),
			Entry("logs_slow_log_loggroup_kms_key_id", "logs_slow_log_loggroup_kms_key_id", "slow-log-key-2"),
			Entry("logs_engine_log_enabled", "logs_engine_log_enabled", true),
			Entry("logs_engine_log_loggroup_retention_in_days", "logs_engine_log_loggroup_retention_in_days", 5),
			Entry("logs_engine_log_loggroup_kms_key_id", "logs_engine_log_loggroup_kms_key_id", "engine-log-key-2"),
			Entry("auto_minor_version_upgrade", "auto_minor_version_upgrade", true),
		)
	})

	Describe("updating node_type for instances of plans which don't specify a fixed node_type", func() {
		It("should allow updating node_type when not enforced by the plan definition", func() {
			instanceID, err := broker.Provision(redisServiceName, "flexible-nodetype-sample", map[string]any{"node_type": "cache.t3.medium"})
			Expect(err).ToNot(HaveOccurred())

			err = broker.Update(instanceID, redisServiceName, "flexible-nodetype-sample", map[string]any{"node_type": "cache.t2.micro"})
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
