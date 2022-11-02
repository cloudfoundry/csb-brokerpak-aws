package terraformtests

import (
	"path"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Aurora mysql", Label("aurora-mysql-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"instance_name":               "csb-auroramysql-test",
		"db_name":                     "csbdb",
		"labels":                      map[string]any{"key1": "some-mysql-value"},
		"region":                      "us-west-2",
		"aws_access_key_id":           awsAccessKeyID,
		"aws_secret_access_key":       awsSecretAccessKey,
		"aws_vpc_id":                  awsVPCID,
		"cluster_instances":           3,
		"serverless_min_capacity":     nil,
		"serverless_max_capacity":     nil,
		"engine_version":              nil,
		"rds_subnet_group":            "",
		"rds_vpc_security_group_ids":  "",
		"allow_major_version_upgrade": true,
		"auto_minor_version_upgrade":  true,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "aurora-mysql/provision")
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
				"engine":                     Equal("aurora-mysql"),
				"identifier":                 Equal("csb-auroramysql-test-0"),
				"instance_class":             Equal("db.r5.large"),
				"db_subnet_group_name":       Equal("csb-auroramysql-test-p-sn"),
				"auto_minor_version_upgrade": BeTrue(),
				"tags":                       HaveKeyWithValue("key1", "some-mysql-value"),
			}))
		})

		It("should create a cluster with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(MatchKeys(IgnoreExtras, Keys{
				"cluster_identifier":                 Equal("csb-auroramysql-test"),
				"engine":                             Equal("aurora-mysql"),
				"database_name":                      Equal("csbdb"),
				"port":                               Equal(float64(3306)),
				"db_subnet_group_name":               Equal("csb-auroramysql-test-p-sn"),
				"skip_final_snapshot":                BeTrue(),
				"serverlessv2_scaling_configuration": BeEmpty(),
				"allow_major_version_upgrade":        BeTrue(),
				"tags":                               HaveKeyWithValue("key1", "some-mysql-value"),
			}))
		})
	})

	When("cluster_instances is 0", func() {
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
				"cluster_identifier":   Equal("csb-auroramysql-test"),
				"engine":               Equal("aurora-mysql"),
				"database_name":        Equal("csbdb"),
				"port":                 Equal(float64(3306)),
				"db_subnet_group_name": Equal("csb-auroramysql-test-p-sn"),
				"skip_final_snapshot":  BeTrue(),
			}))
		})
	})

	When("rds_vpc_security_group_ids is passed", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"rds_vpc_security_group_ids": "group1,group2,group3",
			}))
		})

		It("should use the ids passed and not create new security groups", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(
				MatchKeys(IgnoreExtras, Keys{
					"vpc_security_group_ids": ConsistOf("group1", "group2", "group3"),
				}))
			Expect(ResourceCreationForType(plan, "aws_security_group")).To(BeEmpty())
			Expect(ResourceCreationForType(plan, "aws_security_group_rule")).To(BeEmpty())
		})
	})

	When("rds_subnet_group is passed", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"rds_subnet_group": "some-other-group",
			}))
		})

		It("should use the ids passed and not create new db subnet group", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(
				MatchKeys(IgnoreExtras, Keys{
					"db_subnet_group_name": Equal("some-other-group"),
				}))
			Expect(AfterValuesForType(plan, "aws_rds_cluster_instance")).To(
				MatchKeys(IgnoreExtras, Keys{
					"db_subnet_group_name": Equal("some-other-group"),
				}))

			Expect(ResourceCreationForType(plan, "rds_subnet_group")).To(BeEmpty())
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
				"engine_version": "8.0.mysql_aurora.3.02.0",
			}))
		})

		It("passes the min_capacity, max_capacity and correct instance_class", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(MatchKeys(IgnoreExtras, Keys{
				"engine_version": Equal("8.0.mysql_aurora.3.02.0"),
			}))
		})
	})
})
