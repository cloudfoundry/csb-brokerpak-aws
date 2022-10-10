package terraformtests

import (
	"path"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("mysql", Label("mysql-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"cores":                                 nil,
		"instance_name":                         "csb-mysql-test",
		"db_name":                               "vsbdb",
		"labels":                                map[string]string{"label1": "value1"},
		"storage_gb":                            5,
		"storage_type":                          "io1",
		"iops":                                  3000,
		"publicly_accessible":                   false,
		"multi_az":                              false,
		"instance_class":                        "an-instance-class",
		"engine":                                "mysql",
		"engine_version":                        5.7,
		"aws_vpc_id":                            awsVPCID,
		"storage_autoscale":                     false,
		"storage_autoscale_limit_gb":            0,
		"storage_encrypted":                     false,
		"parameter_group_name":                  "",
		"rds_subnet_group":                      "",
		"rds_vpc_security_group_ids":            "",
		"allow_major_version_upgrade":           true,
		"auto_minor_version_upgrade":            true,
		"maintenance_end_hour":                  nil,
		"maintenance_start_hour":                nil,
		"maintenance_end_min":                   nil,
		"maintenance_start_min":                 nil,
		"maintenance_day":                       nil,
		"use_tls":                               true,
		"deletion_protection":                   false,
		"backup_retention_period":               7,
		"backup_window":                         nil,
		"copy_tags_to_snapshot":                 true,
		"delete_automated_backups":              true,
		"aws_access_key_id":                     awsAccessKeyID,
		"aws_secret_access_key":                 awsSecretAccessKey,
		"region":                                "us-west-2",
		"option_group_name":                     nil,
		"monitoring_interval":                   0,
		"monitoring_role_arn":                   "",
		"performance_insights_enabled":          true,
		"performance_insights_kms_key_id":       "",
		"performance_insights_retention_period": 7,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "mysql/provision")
		Init(terraformProvisionDir)
	})

	Context("mysql parameter groups", func() {
		When("No parameter group name passed", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
			})

			It("should use the default parameter group", func() {
				Expect(ResourceCreationForType(plan, "aws_db_parameter_group")).To(BeEmpty())

				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"parameter_group_name": Equal("default.mysql5.7"),
					}))
			})

		})

		Context("Parameter group passed", func() {
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
		When("default values are passed", func() {
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

		When("storage_type is gp2", func() {
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
		When("no security group ids passed", func() {
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
						"name": Equal("csb-mysql-test-sg"),
					}))
			})
		})

		When("security group ids passed", func() {
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
		When("no window is set", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
			})
			It("should not be passed", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(Not(HaveKey("maintenance_window")))
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
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"maintenance_window": Equal("mon:01:03-mon:02:04"),
					}))
			})

		})

	})

	Context("performance_insights", func() {
		When("is not enabled", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"performance_insights_enabled": false}))
			})
			It("should not set performance_insights_retention_period", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(Not(HaveKey("performance_insights_retention_period")))
			})
		})

		When("is enabled", func() {
			retentionPeriod := 7
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(
					defaultVars,
					map[string]any{
						"performance_insights_enabled":          true,
						"performance_insights_retention_period": retentionPeriod,
					},
				))
			})
			It("should set the passed value for performance_insights_retention_period", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"performance_insights_enabled":          Equal(true),
						"performance_insights_retention_period": BeNumerically("==", retentionPeriod),
					}))
			})
		})
	})

})
