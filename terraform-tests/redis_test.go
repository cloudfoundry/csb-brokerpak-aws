package terraformtests

import (
	"path"

	"github.com/onsi/gomega/gbytes"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Redis", Label("redis-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"cache_size":                         nil,
		"redis_version":                      "6.0",
		"instance_name":                      "csb-redis-test",
		"labels":                             map[string]any{"key1": "some-redis-value"},
		"node_type":                          "cache.t3.medium",
		"node_count":                         2,
		"elasticache_subnet_group":           "",
		"elasticache_vpc_security_group_ids": "",
		"region":                             "us-west-2",
		"aws_access_key_id":                  awsAccessKeyID,
		"aws_secret_access_key":              awsSecretAccessKey,
		"aws_vpc_id":                         awsVPCID,
		"at_rest_encryption_enabled":         true,
		"kms_key_id":                         "fake-encryption-at-rest-key",
		"maintenance_end_hour":               nil,
		"maintenance_start_hour":             nil,
		"maintenance_end_min":                nil,
		"maintenance_start_min":              nil,
		"maintenance_day":                    nil,
		"data_tiering_enabled":               false,
		"multi_az_enabled":                   true,
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
			Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(
				MatchAllKeys(Keys{
					"replication_group_id":       Equal("csb-redis-test"),
					"description":                Equal("csb-redis-test redis"),
					"node_type":                  Equal("cache.t3.medium"),
					"num_cache_clusters":         BeNumerically("==", 2),
					"engine":                     Equal("redis"),
					"engine_version":             Equal("6.0"),
					"port":                       BeNumerically("==", 6379),
					"tags":                       HaveKeyWithValue("key1", "some-redis-value"),
					"subnet_group_name":          Equal("csb-redis-test-p-sn"),
					"transit_encryption_enabled": BeTrue(),
					"automatic_failover_enabled": BeTrue(),
					"apply_immediately":          BeTrue(),
					"at_rest_encryption_enabled": BeTrue(),
					"kms_key_id":                 Equal("fake-encryption-at-rest-key"),

					// By specifying these (apparently less useful) keys in the test we'll
					// get very valuable feedback when bumping the provider (test may break).
					// If a new version adds new properties we will know immediately which
					// will help us stay up-to-date with the provider's latest improvements.
					"notification_topic_arn":      BeNil(),
					"snapshot_name":               BeNil(),
					"snapshot_retention_limit":    BeNil(),
					"timeouts":                    BeNil(),
					"final_snapshot_identifier":   BeNil(),
					"log_delivery_configuration":  BeAssignableToTypeOf([]any{}),
					"availability_zones":          BeNil(),
					"multi_az_enabled":            BeAssignableToTypeOf(false),
					"preferred_cache_cluster_azs": BeNil(),
					"snapshot_arns":               BeNil(),
					"tags_all":                    BeAssignableToTypeOf(map[string]any{}),
					"user_group_ids":              BeNil(),
					"data_tiering_enabled":        BeFalse(),
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
			Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(
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
			Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(
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
			Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(
				MatchKeys(IgnoreExtras, Keys{
					"node_type": Equal("cache.t2.micro"),
				}))
		})
	})

	When("node_count is greater than 1", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"node_count": 2,
			}))
		})

		It("automatic_failover_enabled should be set to true", func() {
			Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(
				MatchKeys(IgnoreExtras, Keys{
					"automatic_failover_enabled": BeTrue(),
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
			Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(MatchKeys(IgnoreExtras, Keys{
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
				Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(Not(HaveKey("maintenance_window")))
			})
		})

		When("maintainance window specified with all values", func() {
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
				Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(
					MatchKeys(IgnoreExtras, Keys{
						"maintenance_window": Equal("mon:01:03-mon:02:04"),
					}))
			})

		})

	})

	Context("multi_az_enabled", func() {
		When("invalid combination", func() {
			It("should not be passed", func() {
				vars := buildVars(defaultVars, map[string]any{"node_count": 1, "multi_az_enabled": true})
				session, err := FailPlan(terraformProvisionDir, vars)

				Expect(session.ExitCode()).NotTo(Equal(0), "Terraform plan should return and error upon exit")
				Expect(session.Err).To(gbytes.Say("automatic_failover_enabled must be true if multi_az_enabled is true"))

				Expect(err).NotTo(HaveOccurred())
			})
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
