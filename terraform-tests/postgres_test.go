package terraformtests

import (
	"path"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

// To execute this test individually: `TF_CLI_CONFIG_FILE="$(pwd)/custom.tfrc" ginkgo --label-filter=postgres-terraform  -v terraform-tests`
var _ = Describe("postgres", Label("postgres-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"instance_name":                         "csb-postgresql-test",
		"db_name":                               "vsbdb",
		"cores":                                 nil,
		"labels":                                map[string]string{"label1": "value1"},
		"storage_gb":                            5,
		"publicly_accessible":                   false,
		"multi_az":                              false,
		"instance_class":                        "db.r5.large",
		"postgres_version":                      14,
		"aws_vpc_id":                            awsVPCID,
		"storage_autoscale":                     false,
		"storage_autoscale_limit_gb":            0,
		"storage_encrypted":                     false,
		"parameter_group_name":                  "",
		"rds_subnet_group":                      "",
		"rds_vpc_security_group_ids":            "",
		"allow_major_version_upgrade":           true,
		"auto_minor_version_upgrade":            false,
		"maintenance_end_hour":                  nil,
		"maintenance_start_hour":                nil,
		"maintenance_end_min":                   nil,
		"maintenance_start_min":                 nil,
		"maintenance_day":                       nil,
		"region":                                "us-west-2",
		"backup_window":                         nil,
		"copy_tags_to_snapshot":                 true,
		"delete_automated_backups":              true,
		"deletion_protection":                   false,
		"iops":                                  3000,
		"kms_key_id":                            "",
		"monitoring_interval":                   0,
		"monitoring_role_arn":                   "",
		"performance_insights_enabled":          false,
		"performance_insights_kms_key_id":       "",
		"performance_insights_retention_period": 7,
		"provider_verify_certificate":           true,
		"require_ssl":                           false,
		"storage_type":                          "io1",
		"backup_retention_period":               7,
		"aws_access_key_id":                     awsAccessKeyID,
		"aws_secret_access_key":                 awsSecretAccessKey,
		"enable_export_postgresql_logs":         false,
		"cloudwatch_postgresql_log_group_retention_in_days": 30,
		"enable_export_upgrade_logs":                        false,
		"cloudwatch_upgrade_log_group_retention_in_days":    30,
		"cloudwatch_log_groups_kms_key_id":                  "",
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "postgresql/provision")
		Init(terraformProvisionDir)
	})

	Context("cloud watch log groups", func() {
		When("no parameters passed", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"postgres_version": "14.1",
				}))
			})

			It("should not create a cloud watch log group", func() {
				Expect(ResourceCreationForType(plan, "aws_cloudwatch_log_group")).To(HaveLen(0))
			})

			It("should not set any cloud watch export", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						// TF provider checks v.(*schema.Set).Len() > 0 to set an array or nil
						"enabled_cloudwatch_logs_exports": BeNil(),
					}),
				)
			})
		})

		When("log groups enabled", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"postgres_version":                                  "14.1",
					"enable_export_postgresql_logs":                     true,
					"enable_export_upgrade_logs":                        true,
					"cloudwatch_postgresql_log_group_retention_in_days": 1,
					"cloudwatch_upgrade_log_group_retention_in_days":    1,
					"cloudwatch_log_groups_kms_key_id":                  "arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx",
				}))
			})

			It("should set two cloud watch export", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"enabled_cloudwatch_logs_exports": ConsistOf("postgresql", "upgrade"),
					}),
				)
			})

			It("should create two cloud watch log groups", func() {
				Expect(GroupAfterValuesForType(plan, "aws_cloudwatch_log_group")).To(
					ConsistOf(
						MatchKeys(IgnoreExtras, Keys{
							"kms_key_id":        Equal("arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx"),
							"name":              Equal("/aws/rds/instance/csb-postgresql-test/postgresql"),
							"retention_in_days": BeNumerically("==", 1),
							"skip_destroy":      BeFalse(),
							"tags":              MatchKeys(IgnoreExtras, Keys{"label1": Equal("value1")}),
						}),
						MatchKeys(IgnoreExtras, Keys{
							"kms_key_id":        Equal("arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx"),
							"name":              Equal("/aws/rds/instance/csb-postgresql-test/upgrade"),
							"retention_in_days": BeNumerically("==", 1),
							"skip_destroy":      BeFalse(),
							"tags":              MatchKeys(IgnoreExtras, Keys{"label1": Equal("value1")}),
						}),
					),
				)
			})
		})

		When("only one log group is enabled", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"postgres_version":                                  "14.1",
					"enable_export_postgresql_logs":                     true,
					"cloudwatch_postgresql_log_group_retention_in_days": 3,
					"cloudwatch_log_groups_kms_key_id":                  "",
				}))
			})

			It("should set one cloud watch export", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"enabled_cloudwatch_logs_exports": ConsistOf("postgresql"),
					}),
				)
			})

			It("should create one cloud watch log group", func() {
				Expect(AfterValuesForType(plan, "aws_cloudwatch_log_group")).To(
					MatchKeys(IgnoreExtras, Keys{
						"kms_key_id":        BeNil(),
						"name":              Equal("/aws/rds/instance/csb-postgresql-test/postgresql"),
						"retention_in_days": BeNumerically("==", 3),
						"skip_destroy":      BeFalse(),
						"tags":              MatchKeys(IgnoreExtras, Keys{"label1": Equal("value1")}),
					}),
				)
			})
		})
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
				// aws_db_instance.parameter_group_name is known after apply. We can't check it.
				Expect(AfterValuesForType(plan, "aws_db_parameter_group")).To(
					MatchKeys(IgnoreExtras, Keys{
						"name_prefix": ContainSubstring("rds-pg-csb-postgresql-test"),
						"family":      Equal("postgres14"),
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
						"name_prefix": ContainSubstring("rds-pg-csb-postgresql-test"),
						"family":      Equal("postgres14"),
						"parameter": ConsistOf(
							MatchKeys(IgnoreExtras, Keys{
								"name":  Equal("rds.force_ssl"),
								"value": Equal("1"),
							}),
						),
					}),
				)
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
						"iops":         BeNumerically("==", 3000),
					}))
			})
		})

		Context("storage_type gp2", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"storage_type": "gp2",
				}))
			})

			It("iops should not be set", func() {
				instanceData := AfterValuesForType(plan, "aws_db_instance")
				Expect(instanceData).To(MatchKeys(IgnoreExtras, Keys{
					"storage_type": Equal("gp2"),
				}))
				Expect("iops").NotTo(BeKeyOf(instanceData))
			})
		})

		When("valid type for iops", func() {
			DescribeTable("iops should be set",
				func(storageTypeParam map[string]any) {
					plan := ShowPlan(terraformProvisionDir, buildVars(defaultVars, storageTypeParam))
					instanceData := AfterValuesForType(plan, "aws_db_instance")

					Expect(instanceData).To(
						MatchKeys(IgnoreExtras, Keys{
							"storage_type": Equal(storageTypeParam["storage_type"]),
						}),
					)
					Expect("iops").To(BeKeyOf(instanceData))
				},
				Entry(
					"io1",
					map[string]any{"storage_type": "io1"},
				),
				Entry(
					"gp3",
					map[string]any{"storage_type": "gp3"},
				),
			)
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

		Context("maintainance window specified with all values", func() {
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

	Context("auto_minor_version_upgrade", func() {
		When("is enabled and a not major version is selected", func() {
			It("should complain about postcondition", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"auto_minor_version_upgrade": true,
					"postgres_version":           14.2,
				}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`Error: Resource postcondition failed`))
				Expect(msgs).To(ContainSubstring(`A Major engine version should be specified when auto_minor_version_upgrade is enabled. Expected postgres version: 14 - got: 14.2`))

				session, _ = FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
					"auto_minor_version_upgrade": true,
					"postgres_version":           14.7,
				}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs = string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`Error: Resource postcondition failed`))
				Expect(msgs).To(ContainSubstring(`A Major engine version should be specified when auto_minor_version_upgrade is enabled. Expected postgres version: 14 - got: 14.7`))
			})
		})
	})

})
