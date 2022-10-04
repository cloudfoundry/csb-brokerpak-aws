package terraformtests

import (
	. "csbbrokerpakaws/terraform-tests/helpers"
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("postgres", Label("postgres-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"instance_name":                   "csb-postgresql-test",
		"db_name":                         "vsbdb",
		"cores":                           nil,
		"labels":                          map[string]string{"label1": "value1"},
		"storage_gb":                      5,
		"publicly_accessible":             false,
		"multi_az":                        false,
		"instance_class":                  "db.r5.large",
		"postgres_version":                14,
		"aws_vpc_id":                      awsVPCID,
		"storage_autoscale":               false,
		"storage_autoscale_limit_gb":      0,
		"storage_encrypted":               false,
		"parameter_group_name":            "",
		"rds_subnet_group":                "",
		"rds_vpc_security_group_ids":      "",
		"allow_major_version_upgrade":     true,
		"auto_minor_version_upgrade":      true,
		"maintenance_end_hour":            nil,
		"maintenance_start_hour":          nil,
		"maintenance_end_min":             nil,
		"maintenance_start_min":           nil,
		"maintenance_day":                 nil,
		"region":                          "us-west-2",
		"backup_window":                   nil,
		"copy_tags_to_snapshot":           true,
		"delete_automated_backups":        true,
		"deletion_protection":             false,
		"iops":                            3000,
		"kms_key_id":                      "",
		"monitoring_interval":             0,
		"monitoring_role_arn":             "",
		"performance_insights_enabled":    false,
		"performance_insights_kms_key_id": "",
		"provider_verify_certificate":     true,
		"require_ssl":                     false,
		"storage_type":                    "io1",
		"backup_retention_period":         7,
		"aws_access_key_id":               awsAccessKeyID,
		"aws_secret_access_key":           awsSecretAccessKey,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "postgresql/provision")
		Init(terraformProvisionDir)
	})

	Context("postgres parameter groups", func() {
		When("no parameter group name passed", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"postgres_version": "14.1",
				}))
			})

			It("should create a parameter group", func() {
				Expect(ResourceCreationForType(plan, "aws_db_parameter_group")).To(HaveLen(1))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"parameter_group_name": Equal("rds-pg-csb-postgresql-test"),
					}))
				Expect(AfterValuesForType(plan, "aws_db_parameter_group")).To(
					MatchKeys(IgnoreExtras, Keys{
						"name":   Equal("rds-pg-csb-postgresql-test"),
						"family": Equal("postgres14"),
						"parameter": ConsistOf(MatchKeys(IgnoreExtras, Keys{
							"name":  Equal("rds.force_ssl"),
							"value": Equal("0"),
						}))}))
			})
		})

		When("requiring SSL", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"require_ssl": true,
				}))
			})

			It("should configure the parameter in the parameter group", func() {
				Expect(AfterValuesForType(plan, "aws_db_parameter_group")).To(
					MatchKeys(IgnoreExtras, Keys{
						"parameter": ConsistOf(MatchKeys(IgnoreExtras, Keys{
							"name":  Equal("rds.force_ssl"),
							"value": Equal("1"),
						}))}))
			})
		})

		When("parameter group passed", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"parameter_group_name": "some-parameter-group-name",
				}))
			})

			It("should not create a parameter group if name is provided", func() {
				Expect(ResourceCreationForType(plan, "aws_db_parameter_group")).To(BeEmpty())

				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"parameter_group_name": Equal("some-parameter-group-name"),
					}))
			})
		})

	})

	Context("storage type", func() {
		Context("default values", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
			})

			It("default values work with io1 and 3000 iops", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"storage_type": Equal("io1"),
						"iops":         Equal(float64(3000)),
					}))
			})
		})

		Context("storage_type gp2", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"storage_type": "gp2",
				}))
			})

			It("iops should be null", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"storage_type": Equal("gp2"),
						"iops":         BeNil(),
					}))
			})
		})
	})

	Context("autoscaling", func() {
		When("storage_autoscale is false", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"storage_autoscale":          false,
					"storage_autoscale_limit_gb": 200,
				}))
			})

			It("autoscaling should be disabled", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"max_allocated_storage": BeNil(),
					}))
			})
		})

		When("storage_autoscale is true and limit > storage_gb", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"storage_autoscale":          true,
					"storage_autoscale_limit_gb": 200,
				}))
			})

			It("autoscaling should be enabled", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"max_allocated_storage": Equal(float64(200)),
					}))
			})
		})

		When("storage_autoscale is true and limit <= storage_gb", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"storage_autoscale":          true,
					"storage_autoscale_limit_gb": 5,
				}))
			})

			It("autoscaling should be disabled", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"max_allocated_storage": BeNil(),
					}))
			})
		})
	})

	Context("security groups", func() {
		Context("no security group ids passed", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
			})

			It("should create a new one", func() {
				Expect(UnknownValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"vpc_security_group_ids": BeTrue(),
					}))
				Expect(ResourceCreationForType(plan, "aws_security_group")).To(HaveLen(1))
				Expect(AfterValuesForType(plan, "aws_security_group")).To(
					MatchKeys(IgnoreExtras, Keys{
						"name": Equal("csb-postgresql-test-sg"),
					}))
			})
		})

		Context("security group ids passed", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"rds_vpc_security_group_ids": "group1,group2,group3",
				}))
			})

			It("should use the ids passed and no create new security groups", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"vpc_security_group_ids": ConsistOf("group1", "group2", "group3"),
					}))
				Expect(ResourceCreationForType(plan, "aws_security_group")).To(BeEmpty())
			})
		})
	})

	Context("maintenance_window", func() {
		Context("no window", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
			})

			It("should not be passed", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(Not(HaveKey("maintenance_window")))
			})
		})

		Context("only maintenance_day specified", func() {
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
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"maintenance_window": Equal("mon:01:03-mon:02:04"),
					}))
			})
		})
	})
})
