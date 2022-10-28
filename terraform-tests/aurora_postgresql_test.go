package terraformtests

import (
	"path"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Aurora postgresql", Label("aurora-postgresql-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"instance_name":               "csb-aurorapg-test",
		"db_name":                     "csbdb",
		"region":                      "us-west-2",
		"aws_access_key_id":           awsAccessKeyID,
		"aws_secret_access_key":       awsSecretAccessKey,
		"aws_vpc_id":                  awsVPCID,
		"cluster_instances":           3,
		"serverless_min_capacity":     nil,
		"serverless_max_capacity":     nil,
		"engine_version":              nil,
		"allow_major_version_upgrade": true,
		"auto_minor_version_upgrade":  true,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "aurora-postgresql/provision")
		Init(terraformProvisionDir)
	})

	Context("with Default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(9))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_rds_cluster_instance",
				"aws_rds_cluster_instance",
				"aws_rds_cluster_instance",
				"aws_rds_cluster",
				"random_password",
				"random_string",
				"aws_security_group_rule",
				"aws_db_subnet_group",
				"aws_security_group",
			))
		})

		It("should create a cluster_instance with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster_instance")).To(MatchKeys(IgnoreExtras, Keys{
				"engine":                     Equal("aurora-postgresql"),
				"identifier":                 Equal("csb-aurorapg-test-0"),
				"instance_class":             Equal("db.r5.large"),
				"db_subnet_group_name":       Equal("csb-aurorapg-test-p-sn"),
				"auto_minor_version_upgrade": BeTrue(),
			}))
		})

		It("should create a cluster with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(MatchKeys(IgnoreExtras, Keys{
				"cluster_identifier":                 Equal("csb-aurorapg-test"),
				"engine":                             Equal("aurora-postgresql"),
				"database_name":                      Equal("csbdb"),
				"port":                               BeNumerically("==", 5432),
				"db_subnet_group_name":               Equal("csb-aurorapg-test-p-sn"),
				"skip_final_snapshot":                BeTrue(),
				"serverlessv2_scaling_configuration": BeEmpty(),
				"allow_major_version_upgrade":        BeTrue(),
			}))
		})
	})

	Context("cluster_instances is 0", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"cluster_instances": 0,
			}))
		})

		It("should not create any cluster_instance", func() {
			Expect(plan.ResourceChanges).To(HaveLen(6))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_rds_cluster",
				"random_password",
				"random_string",
				"aws_security_group_rule",
				"aws_db_subnet_group",
				"aws_security_group",
			))
		})

		It("should create a cluster with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(MatchKeys(IgnoreExtras, Keys{
				"cluster_identifier":   Equal("csb-aurorapg-test"),
				"engine":               Equal("aurora-postgresql"),
				"database_name":        Equal("csbdb"),
				"port":                 BeNumerically("==", 5432),
				"db_subnet_group_name": Equal("csb-aurorapg-test-p-sn"),
				"skip_final_snapshot":  BeTrue(),
			}))
		})
	})

	Context("serverless", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"serverless_min_capacity": 0.5,
				"serverless_max_capacity": 11.0,
			}))
		})

		It("passes the min_capacity, max_capacity and correct instance_class", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(MatchKeys(IgnoreExtras, Keys{
				"serverlessv2_scaling_configuration": ConsistOf(MatchAllKeys(Keys{
					"min_capacity": Equal(0.5),
					"max_capacity": Equal(11.0),
				})),
			}))

			Expect(AfterValuesForType(plan, "aws_rds_cluster_instance")).To(MatchKeys(IgnoreExtras, Keys{
				"instance_class": Equal("db.serverless"),
			}))
		})
	})

	Context("engine_version", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"engine_version": "8.0.postgresql_aurora.3.02.0",
			}))
		})

		It("passes the min_capacity, max_capacity and correct instance_class", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(MatchKeys(IgnoreExtras, Keys{
				"engine_version": Equal("8.0.postgresql_aurora.3.02.0"),
			}))
		})
	})
})
