package terraformtests

import (
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "csbbrokerpakaws/terraform-tests/helpers"
)

var _ = Describe("Redis", Label("redis-terraform"), Ordered, func() {
	const resource = "aws_elasticache_replication_group"

	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"cache_size":                                 nil,
		"redis_version":                              "6.0",
		"instance_name":                              "csb-redis-test",
		"labels":                                     map[string]any{"key1": "some-redis-value"},
		"node_type":                                  "cache.t3.medium",
		"node_count":                                 2,
		"elasticache_subnet_group":                   "",
		"elasticache_vpc_security_group_ids":         "",
		"region":                                     "us-west-2",
		"aws_access_key_id":                          awsAccessKeyID,
		"aws_secret_access_key":                      awsSecretAccessKey,
		"aws_vpc_id":                                 awsVPCID,
		"at_rest_encryption_enabled":                 true,
		"kms_key_id":                                 "fake-encryption-at-rest-key",
		"maintenance_end_hour":                       nil,
		"maintenance_start_hour":                     nil,
		"maintenance_end_min":                        nil,
		"maintenance_start_min":                      nil,
		"maintenance_day":                            nil,
		"data_tiering_enabled":                       false,
		"automatic_failover_enabled":                 true,
		"multi_az_enabled":                           true,
		"backup_retention_limit":                     12,
		"final_backup_identifier":                    "tortoise",
		"backup_name":                                "turtle",
		"backup_end_hour":                            nil,
		"backup_start_hour":                          nil,
		"backup_end_min":                             nil,
		"backup_start_min":                           nil,
		"parameter_group_name":                       "fake-param-group-name",
		"preferred_azs":                              nil,
		"logs_slow_log_loggroup_kms_key_id":          "",
		"logs_slow_log_loggroup_retention_in_days":   0,
		"logs_slow_log_enabled":                      false,
		"logs_engine_log_loggroup_kms_key_id":        "",
		"logs_engine_log_loggroup_retention_in_days": 0,
		"logs_engine_log_enabled":                    false,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "redis/cluster/provision")
		Init(terraformProvisionDir)
	})

	Context("with Default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(ResourceChangesTypes(plan)).To(ConsistOf(getExpectedResources()))
		})

		It("should create a aws_elasticache_replication_group with the right values", func() {
			Expect(AfterValuesForType(plan, resource)).To(
				MatchAllKeys(Keys{
					"replication_group_id":        Equal("csb-redis-test"),
					"description":                 Equal("csb-redis-test redis"),
					"node_type":                   Equal("cache.t3.medium"),
					"num_cache_clusters":          BeNumerically("==", 2),
					"engine":                      Equal("redis"),
					"engine_version":              Equal("6.0"),
					"port":                        BeNumerically("==", 6379),
					"tags":                        HaveKeyWithValue("key1", "some-redis-value"),
					"subnet_group_name":           Equal("csb-redis-test-p-sn"),
					"transit_encryption_enabled":  BeTrue(),
					"automatic_failover_enabled":  BeTrue(),
					"apply_immediately":           BeTrue(),
					"at_rest_encryption_enabled":  BeTrue(),
					"kms_key_id":                  Equal("fake-encryption-at-rest-key"),
					"snapshot_retention_limit":    BeNumerically("==", 12),
					"final_snapshot_identifier":   Equal("tortoise"),
					"snapshot_name":               Equal("turtle"),
					"auto_minor_version_upgrade":  Equal("false"), // yes, a string. Provider quirk.
					"parameter_group_name":        Equal("fake-param-group-name"),
					"preferred_cache_cluster_azs": BeNil(),

					// By specifying these (apparently less useful) keys in the test we'll
					// get very valuable feedback when bumping the provider (test may break).
					// If a new version adds new properties we will know immediately which
					// will help us stay up-to-date with the provider's latest improvements.
					"notification_topic_arn":     BeNil(),
					"timeouts":                   BeNil(),
					"log_delivery_configuration": BeAssignableToTypeOf([]any{}),
					"availability_zones":         BeNil(),
					"multi_az_enabled":           BeAssignableToTypeOf(false),
					"snapshot_arns":              BeNil(),
					"tags_all":                   BeAssignableToTypeOf(map[string]any{}),
					"user_group_ids":             BeNil(),
					"data_tiering_enabled":       BeFalse(),
				}))
		})
	})

	When("elasticache_vpc_security_group_ids is passed", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"elasticache_vpc_security_group_ids": "group1,group2,group3",
			}))
		})

		It("should not create any security groups or rules", func() {
			nosecuriryGroupsOrRules := Filter(getExpectedResources(), "aws_security_group", "aws_security_group_rule")
			Expect(ResourceChangesTypes(plan)).To(ConsistOf(nosecuriryGroupsOrRules))
		})

		It("should use the elasticache_vpc_security_group_ids passed as the security_group_ids", func() {
			Expect(AfterValuesForType(plan, resource)).To(
				MatchKeys(IgnoreExtras, Keys{
					"security_group_ids": ConsistOf("group1", "group2", "group3"),
				}))
		})
	})

	When("elasticache_subnet_group is passed", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"elasticache_subnet_group": "some-other-group",
			}))
		})

		It("should not create any subnet group", func() {
			noSubnetGroup := Filter(getExpectedResources(), "aws_elasticache_subnet_group")
			Expect(ResourceChangesTypes(plan)).To(ConsistOf(noSubnetGroup))
		})

		It("should use the elasticache_subnet_group passed as the subnet_group_name", func() {
			Expect(AfterValuesForType(plan, resource)).To(
				MatchKeys(IgnoreExtras, Keys{
					"subnet_group_name": Equal("some-other-group"),
				}))
		})
	})

	When("node_type is not empty", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"node_type": "cache.t2.micro",
			}))
		})

		It("should ignore cache_size and create a aws_elasticache_replication_group with that node_type", func() {
			Expect(AfterValuesForType(plan, resource)).To(
				MatchKeys(IgnoreExtras, Keys{
					"node_type": Equal("cache.t2.micro"),
				}))
		})
	})

	Context("redis_version is passed", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"redis_version": "5.0.6",
			}))
		})

		It("should create a aws_elasticache_replication_group with that engine_version", func() {
			Expect(AfterValuesForType(plan, resource)).To(MatchKeys(IgnoreExtras, Keys{
				"engine_version": Equal("5.0.6"),
			}))
		})
	})

	Context("maintenance_window", func() {
		When("no window is set", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
			})

			It("should not be passed", func() {
				Expect(AfterValuesForType(plan, resource)).To(Not(HaveKey("maintenance_window")))
			})
		})

		When("maintenance window specified with all values", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"maintenance_day":        "Mon",
					"maintenance_start_hour": "01",
					"maintenance_end_hour":   "02",
					"maintenance_start_min":  "03",
					"maintenance_end_min":    "04",
				}))
			})

			It("should pass the correct window", func() {
				Expect(AfterValuesForType(plan, resource)).To(
					MatchKeys(IgnoreExtras, Keys{
						"maintenance_window": Equal("mon:01:03-mon:02:04"),
					}))
			})
		})
	})

	Context("backup_window", func() {
		When("no window is set", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
			})

			It("should not be passed", func() {
				Expect(AfterValuesForType(plan, resource)).To(Not(HaveKey("backup_window")))
			})
		})

		When("backup window specified with all values", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"backup_start_hour": "01",
					"backup_end_hour":   "02",
					"backup_start_min":  "03",
					"backup_end_min":    "04",
				}))
			})

			It("should pass the correct window", func() {
				Expect(AfterValuesForType(plan, resource)).To(
					MatchKeys(IgnoreExtras, Keys{
						"snapshot_window": Equal("01:03-02:04"),
					}))
			})
		})
	})

	Context("preferred_azs are passed", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"preferred_azs": []string{"fake-az1", "fake-az2"},
			}))
		})

		It("should create a aws_elasticache_replication_group with that engine_version", func() {
			Expect(AfterValuesForType(plan, resource)).To(MatchKeys(IgnoreExtras, Keys{
				"preferred_cache_cluster_azs": ConsistOf("fake-az1", "fake-az2"),
			}))
		})
	})

	Context("slow_log is enabled", func() {
		var kmsKeyID string
		var retentionInDays int

		BeforeEach(func() {
			kmsKeyID = ""
			retentionInDays = 0
		})

		JustBeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"logs_slow_log_loggroup_kms_key_id":        kmsKeyID,
				"logs_slow_log_loggroup_retention_in_days": retentionInDays,
				"logs_slow_log_enabled":                    true,
			}))
		})

		It("should configure a log group for slow log with default values", func() {
			slowLogResource := "aws_cloudwatch_log_group"

			By("creating a new loggroup resouce")
			expectedResources := []string{slowLogResource}
			expectedResources = append(expectedResources, getExpectedResources()...)
			Expect(ResourceChangesTypes(plan)).To(ConsistOf(expectedResources))

			By("configuring the log group with the default values")
			Expect(AfterValuesForType(plan, slowLogResource)).To(MatchKeys(IgnoreExtras, Keys{
				"name":              Equal("/aws/elasticache/cluster/csb-redis-test/slow-log"),
				"retention_in_days": BeNumerically("==", 0),
				"kms_key_id":        BeNil(),
				"skip_destroy":      BeFalse(),
				"tags":              HaveKeyWithValue("key1", "some-redis-value"),
			}))

			By("assignining that log group to the replication group")
			Expect(AfterValuesForType(plan, resource)).To(MatchKeys(IgnoreExtras, Keys{
				"log_delivery_configuration": ConsistOf(MatchAllKeys(Keys{
					"destination":      Equal("/aws/elasticache/cluster/csb-redis-test/slow-log"),
					"destination_type": Equal("cloudwatch-logs"),
					"log_format":       Equal("json"),
					"log_type":         Equal("slow-log"),
				})),
			}))

		})

		Context("slow_log loggroup custom configuration", func() {
			BeforeEach(func() {
				kmsKeyID = "test-kms-key"
				retentionInDays = 180
			})

			It("should configure the passed values", func() {
				slowLogResource := "aws_cloudwatch_log_group"
				Expect(AfterValuesForType(plan, slowLogResource)).To(MatchKeys(IgnoreExtras, Keys{
					"name":              Equal("/aws/elasticache/cluster/csb-redis-test/slow-log"),
					"retention_in_days": BeNumerically("==", retentionInDays),
					"kms_key_id":        Equal(kmsKeyID),
					"skip_destroy":      BeFalse(),
					"tags":              HaveKeyWithValue("key1", "some-redis-value"),
				}))
			})
		})
	})

	Context("engine_log is enabled", func() {
		var kmsKeyID string
		var retentionInDays int

		BeforeEach(func() {
			kmsKeyID = ""
			retentionInDays = 0
		})

		JustBeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"logs_engine_log_loggroup_kms_key_id":        kmsKeyID,
				"logs_engine_log_loggroup_retention_in_days": retentionInDays,
				"logs_engine_log_enabled":                    true,
			}))
		})

		It("should configure a log group for engine log with default values", func() {
			engineLogResource := "aws_cloudwatch_log_group"

			By("creating a new loggroup resouce")
			expectedResources := []string{engineLogResource}
			expectedResources = append(expectedResources, getExpectedResources()...)
			Expect(ResourceChangesTypes(plan)).To(ConsistOf(expectedResources))

			By("configuring the log group with the default values")
			Expect(AfterValuesForType(plan, engineLogResource)).To(MatchKeys(IgnoreExtras, Keys{
				"name":              Equal("/aws/elasticache/cluster/csb-redis-test/engine-log"),
				"retention_in_days": BeNumerically("==", 0),
				"kms_key_id":        BeNil(),
				"skip_destroy":      BeFalse(),
				"tags":              HaveKeyWithValue("key1", "some-redis-value"),
			}))

			By("assignining that log group to the replication group")
			Expect(AfterValuesForType(plan, resource)).To(MatchKeys(IgnoreExtras, Keys{
				"log_delivery_configuration": ConsistOf(MatchAllKeys(Keys{
					"destination":      Equal("/aws/elasticache/cluster/csb-redis-test/engine-log"),
					"destination_type": Equal("cloudwatch-logs"),
					"log_format":       Equal("json"),
					"log_type":         Equal("engine-log"),
				})),
			}))
		})

		Context("engine loggroup custom coonfiguration", func() {
			BeforeEach(func() {
				kmsKeyID = "test-kms-key-2"
				retentionInDays = 5
			})

			It("should configure the specified values", func() {
				engineLogResource := "aws_cloudwatch_log_group"
				Expect(AfterValuesForType(plan, engineLogResource)).To(MatchKeys(IgnoreExtras, Keys{
					"name":              Equal("/aws/elasticache/cluster/csb-redis-test/engine-log"),
					"retention_in_days": BeNumerically("==", 5),
					"kms_key_id":        Equal("test-kms-key-2"),
					"skip_destroy":      BeFalse(),
					"tags":              HaveKeyWithValue("key1", "some-redis-value"),
				}))
			})
		})
	})

	Context("node_count is 1", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"node_count": 1,
			}))
		})

		It("sets `automatic_failover_enabled` and `multi_az_enabled` to false", func() {
			Expect(AfterValuesForType(plan, resource)).To(MatchKeys(IgnoreExtras, Keys{
				"automatic_failover_enabled": BeFalse(),
				"multi_az_enabled":           BeFalse(),
			}))
		})
	})
})

func getExpectedResources() []string {
	// This tries to be equivalent to a constant slice.
	// if it was a variable it could be changed accidentally.
	return []string{
		"aws_elasticache_replication_group",
		"random_password",
		"aws_security_group",
		"aws_elasticache_subnet_group",
		"aws_security_group_rule",
	}
}
