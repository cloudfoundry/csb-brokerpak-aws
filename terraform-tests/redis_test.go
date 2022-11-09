package terraformtests

import (
	"path"

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
		"cache_size":                         2,
		"redis_version":                      "6.0",
		"instance_name":                      "csb-redis-test",
		"labels":                             map[string]any{"key1": "some-redis-value"},
		"node_type":                          "",
		"node_count":                         1,
		"elasticache_subnet_group":           "",
		"elasticache_vpc_security_group_ids": "",
		"region":                             "us-west-2",
		"aws_access_key_id":                  awsAccessKeyID,
		"aws_secret_access_key":              awsSecretAccessKey,
		"aws_vpc_id":                         awsVPCID,
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
			Expect(plan.ResourceChanges).To(HaveLen(5))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_elasticache_replication_group",
				"random_password",
				"aws_security_group",
				"aws_elasticache_subnet_group",
				"aws_security_group_rule",
			))
		})

		It("should create a aws_elasticache_replication_group with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_elasticache_replication_group")).To(
				MatchKeys(IgnoreExtras, Keys{
					"replication_group_id":       Equal("csb-redis-test"),
					"description":                Equal("csb-redis-test redis"),
					"node_type":                  Equal("cache.t3.medium"),
					"num_cache_clusters":         BeNumerically("==", 1),
					"engine":                     Equal("redis"),
					"engine_version":             Equal("6.0"),
					"port":                       BeNumerically("==", 6379),
					"tags":                       HaveKeyWithValue("key1", "some-redis-value"),
					"subnet_group_name":          Equal("csb-redis-test-p-sn"),
					"transit_encryption_enabled": BeTrue(),
					"automatic_failover_enabled": BeFalse(),
				}))
		})
	})
})
