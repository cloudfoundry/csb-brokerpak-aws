package terraformtests

import (
	"path"

	"golang.org/x/exp/maps"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("mssql", Label("mssql-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"aws_access_key_id":     awsAccessKeyID,
		"aws_secret_access_key": awsSecretAccessKey,
		"region":                "us-west-2",
		"instance_name":         "csb-mssql-test",
		"storage_encrypted":     true,
		"kms_key_id":            "",
		"db_name":               "vsbdb",
		"labels":                map[string]string{"label1": "value1"},
		"max_allocated_storage": 0,
		"deletion_protection":   true,
		"publicly_accessible":   false,

		"aws_vpc_id":                 "",
		"rds_subnet_group":           "",
		"rds_vpc_security_group_ids": "",
		"option_group_name":          "",
		"parameter_group_name":       "",

		"storage_type": "io1",
		"iops":         3000,
		"multi_az":     true,

		"backup_window":            nil,
		"copy_tags_to_snapshot":    true,
		"backup_retention_period":  7,
		"delete_automated_backups": true,
		"maintenance_end_hour":     nil,
		"maintenance_start_hour":   nil,
		"maintenance_end_min":      nil,
		"maintenance_start_min":    nil,
		"maintenance_day":          nil,
		"character_set_name":       nil,

		"allow_major_version_upgrade": true,
		"auto_minor_version_upgrade":  true,

		"performance_insights_enabled":          false,
		"performance_insights_kms_key_id":       "",
		"performance_insights_retention_period": 7,
	}

	requiredVars := map[string]any{
		"engine":        "sqlserver-ee",
		"mssql_version": "15.00",
		"storage_gb":    20,

		"instance_class": "some-instance-class",

		"monitoring_interval": 0,
		"monitoring_role_arn": "",
	}

	validVPC := awsVPCID

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "mssql/provision")
		Init(terraformProvisionDir)
	})

	Context("with Default and required values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(11))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_db_instance",
				"random_password",
				"random_string",
				"aws_security_group_rule",
				"aws_security_group_rule",
				"aws_security_group_rule",
				"aws_security_group_rule",
				"aws_security_group_rule",
				"aws_db_subnet_group",
				"aws_security_group",
				"aws_db_parameter_group",
			))
		})

		It("should create a db instance with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{
				"engine":               Equal("sqlserver-ee"),
				"engine_version":       Equal("15.00"),
				"identifier":           Equal("csb-mssql-test"),
				"storage_encrypted":    BeTrue(),
				"instance_class":       Equal("some-instance-class"),
				"tags":                 HaveKeyWithValue("label1", "value1"),
				"db_subnet_group_name": Equal("csb-mssql-test-p-sn"),
				"apply_immediately":    BeTrue(),
				"skip_final_snapshot":  BeTrue(),
				"license_model":        Equal("license-included"),
				"publicly_accessible":  BeFalse(),
				"monitoring_interval":  BeNumerically("==", 0),

				"performance_insights_enabled": BeFalse(),
			}))
		})
	})

	Context("with Default values alone", func() {
		It("should complain about missing required values", func() {
			session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars))
			Expect(session.ExitCode()).NotTo(Equal(0))
			msgs := string(session.Out.Contents())
			Expect(msgs).To(ContainSubstring(`The root module input variable \"mssql_version\" is not set, and has no default value.`))
			Expect(msgs).To(ContainSubstring(`The root module input variable \"engine\" is not set, and has no default value.`))
			Expect(msgs).To(ContainSubstring(`The root module input variable \"storage_gb\" is not set, and has no default value.`))
			Expect(msgs).To(ContainSubstring(`The root module input variable \"instance_class\" is not set, and has no default value.`))
		})
	})

	Context("simple pass-through fields for aws_db_instance", func() {
		DescribeTable("are passed to Terraform aws_db_instance unmodified",
			func(propName string, propValue any) {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{propName: propValue}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{propName: Equal(propValue)}))
			},
			Entry("passthrough", "character_set_name", "a_custom_charset_value"),
			Entry("passthrough", "multi_az", false),
		)
	})

	Context("monitoring_interval", func() {
		When("monitoring_role_arn is invalid", func() {
			It("complains about invalid monitoring_role_arn", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"monitoring_role_arn": "NOTVALID"}))
				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`(NOTVALID) is an invalid ARN: arn: invalid prefix`))
			})
		})

		When("monitoring_role_arn has a valid prefix but invalid account id", func() {
			It("complains about invalid account id value", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"monitoring_role_arn": "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access"}))
				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`(arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access) is an invalid ARN: invalid account ID value (expecting to match regular expression: ^(aws|aws-managed|third-party|\\d{12}|cw.{10})$)`))
			})
		})

		When("monitoring_interval is invalid", func() {
			It("complains about invalid monitoring_interval", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"monitoring_interval": -1}))
				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`expected monitoring_interval to be one of [0 1 5 10 15 30 60], got -1`))
			})
		})

		When("monitoring_interval is invalid in positive range", func() {
			It("complains about invalid monitoring_interval", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"monitoring_interval": 9}))
				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`expected monitoring_interval to be one of [0 1 5 10 15 30 60], got 9`))
			})
		})

		When("monitoring_interval and monitoring_role_arn are valid", func() {
			It("succeeds to plan operations", func() {
				plan := ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"monitoring_interval": 60, "monitoring_role_arn": "arn:aws:iam::123456789012:test"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"monitoring_interval": BeNumerically("==", 60), "monitoring_role_arn": Equal("arn:aws:iam::123456789012:test")}))
			})
		})
	})

	Context("csbmajorengineversion provider needs a valid engine version", func() {
		When("mssql_version is not valid", func() {
			It("it fails when recovering the major version when creating the db parameter group", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"mssql_version": "ANY-VALUE-AT-ALL"}))
				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`invalid parameter combination. API does not return any db engine version - engine sqlserver-ee - engine version ANY-VALUE-AT-ALL`))
			})
		})
	})

	Context("properties that pass-through without validation", func() {
		When("instance_class is passed", func() {
			It("it passes-through without validation", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"instance_class": "ANY-VALUE-AT-ALL"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"instance_class": Equal("ANY-VALUE-AT-ALL")}))
			})
		})
	})
	Context("instance_name", func() {
		When("invalid instance_name is passed", func() {
			It("fails and returns a descriptive message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"instance_name": "THIS-ENGINE-DOESNT-EXIST"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session).To(gbytes.Say(`only lowercase alphanumeric characters, hyphens, underscores, periods, and spaces allowed in \\"name\\"`))

			})
		})
		When("instance_name is passed", func() {
			It("it is used to populate several fields", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"instance_name": "some-better-value"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"identifier": Equal("some-better-value")}))
				Expect(AfterValuesForType(plan, "aws_security_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("some-better-value-sg")}))
				Expect(AfterValuesForType(plan, "aws_db_subnet_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("some-better-value-p-sn")}))
			})
		})
	})

	Context("engine", func() {
		When("valid engine passed", func() {
			It("should use the passed value", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"engine": "sqlserver-web"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"engine": Equal("sqlserver-web")}))
			})
		})
	})

	Context("storage_gb", func() {
		When("valid storage_gb passed", func() {
			It("should use the passed value", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"storage_gb": 60}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"allocated_storage": Equal(float64(60))}))
			})
		})
	})

	Context("max_allocated_storage", func() {
		When("valid max_allocated_storage passed", func() {
			It("should use the passed value", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"max_allocated_storage": 250}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"max_allocated_storage": Equal(float64(250))}))
			})
		})
	})

	Context("db_name", func() {
		When("passed", func() {
			It("is not assigned to the db_instance", func() {
				// This behaviour is intentional. AWS RDS for SQLServer doesn't allow specifying a db_name.
				// The db_name will be passed to the binding's provider which will be responsible for creating the db.
				// https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"db_name": "ANY-VALUE-AT-ALL"}))
				Expect(AfterValuesForType(plan, "")).To(BeNil())
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras|IgnoreMissing, Keys{"db_name": BeNil()}))
			})
		})
	})

	Context("aws_vpc_id", func() {
		When("no vpc passed", func() {
			It("should succeed and use the default one", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, nil))
				Expect(session.ExitCode()).To(Equal(0))
			})
		})

		When("valid vpc passed", func() {
			It("should create the right resources", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"aws_vpc_id": awsVPCID}))

				Expect(AfterValuesForType(plan, "aws_security_group")).To(MatchKeys(IgnoreExtras, Keys{"vpc_id": Equal(awsVPCID)}))

				Expect(ResourceChangesTypes(plan)).To(ConsistOf(
					"aws_db_instance",
					"random_password",
					"random_string",
					"aws_security_group_rule",
					"aws_security_group_rule",
					"aws_security_group_rule",
					"aws_security_group_rule",
					"aws_security_group_rule",
					"aws_db_subnet_group",
					"aws_security_group",
					"aws_db_parameter_group",
				))
			})
		})

		When("invalid vpc passed", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"aws_vpc_id": "THIS-VPC-DOESNT-EXIST"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session).To(gbytes.Say("no matching EC2 VPC found"))
			})
		})
	})

	Context("subnet group", func() {
		When("no subnet group passed", func() {
			It("should create a new one", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, nil))
				Expect(ResourceCreationForType(plan, "aws_db_subnet_group")).To(HaveLen(1))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"db_subnet_group_name": Equal("csb-mssql-test-p-sn")}))
				Expect(AfterValuesForType(plan, "aws_db_subnet_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("csb-mssql-test-p-sn")}))
			})
		})

		When("a subnet group passed without specifying a vpc", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"rds_subnet_group": "ANY-SUBNET-GROUP"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session).To(gbytes.Say("when specifying rds_subnet_group please specify also the corresponding aws_vpc_id"))
			})
		})

		When("invalid subnet group passed", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"rds_subnet_group": "THIS-SUBNET-GROUP-DOESNT-EXIST", "aws_vpc_id": validVPC}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session).To(gbytes.Say(`no matching RDS DB Subnet Group found`))
			})
		})
	})

	Context("security groups", func() {
		When("no security group ids passed and no vpc passed", func() {
			It("should create a new security group in the default vpc", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars))
				Expect(ResourceCreationForType(plan, "aws_security_group")).To(HaveLen(1))
				Expect(UnknownValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"vpc_security_group_ids": BeTrue()}))
				Expect(AfterValuesForType(plan, "aws_security_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("csb-mssql-test-sg")}))
			})
		})

		When("no security group ids passed and a valid vpc passed", func() {
			It("should create a new security group in the specified vpc", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"aws_vpc_id": validVPC}))
				Expect(ResourceCreationForType(plan, "aws_security_group")).To(HaveLen(1))
				Expect(UnknownValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"vpc_security_group_ids": BeTrue()}))
				Expect(AfterValuesForType(plan, "aws_security_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("csb-mssql-test-sg"), "vpc_id": Equal(validVPC)}))
			})
		})

		When("a security group ids passed without specifying a vpc", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"rds_vpc_security_group_ids": "ANY,SECURITY,GROUP"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session).To(gbytes.Say(`when specifying rds_vpc_security_group_ids please specify also the corresponding aws_vpc_id`))
			})
		})

		When("invalid security group ids passed", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"rds_vpc_security_group_ids": "THESE,SECURITY-GROUPS,DONT-EXIST", "aws_vpc_id": validVPC}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session).To(gbytes.Say(`the specified security groups don't exist or don't correspond to the specified vpc \(1\)`))
			})
		})
	})

	Context("kms_key_id", func() {
		When("with default values", func() {
			It("should succeed", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, nil))
				Expect(ResourceCreationForType(plan, "aws_db_subnet_group")).To(HaveLen(1))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"storage_encrypted": BeTrue()}))
			})
		})

		When("no kms_key_id passed and storage_encrypted set to true", func() {
			// Currently, this tests the exact same inputs as the one `with default values`
			// However, I believe it is useful to keep it in case we decide to change the defaults
			It("should succeed and use a kms key managed by AWS to encrypt the db", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"kms_key_id": "", "storage_encrypted": true}))
				Expect(ResourceCreationForType(plan, "aws_db_subnet_group")).To(HaveLen(1))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"storage_encrypted": BeTrue()}))
			})
		})

		When("a kms_key_id is passed and storage_encrypted is false", func() {
			It("should complain about kms_key_id and storage_encrypted mismatch - storage_encrypted: false", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"kms_key_id": "some-kms-id", "storage_encrypted": false}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session).To(gbytes.Say("set `storage_encrypted` to `true` or leave `kms_key_id` field blank"))
			})
		})

		When("an invalid kms_key_id is passed and storage_encrypted is true", func() {
			It("should complain about kms_key_id not having a valid syntax", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"kms_key_id": "some-kms-id", "storage_encrypted": true}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session).To(gbytes.Say("is an invalid ARN: arn: invalid prefix"))
			})
		})

		When("no kms_key_id passed and storage_encrypted set to false", func() {
			It("should proceed without issues", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"kms_key_id": "", "storage_encrypted": false}))
				Expect(ResourceCreationForType(plan, "aws_db_subnet_group")).To(HaveLen(1))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"storage_encrypted": BeFalse()}))
			})
		})
	})

	Context("storage type", func() {
		Context("default values", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{}))
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
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
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
					plan := ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, storageTypeParam))
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

	Context("sqlserver validations", func() {
		DescribeTable("sqlserver-ex is the only edition without encryption support",
			func(extraProps map[string]any, expectedError string) {
				if expectedError == "" {
					plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, extraProps))
				} else {
					session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, extraProps))
					Expect(session.ExitCode()).NotTo(Equal(0))
					Expect(session).To(gbytes.Say(expectedError))
				}
			},
			Entry("sqlserver-ex  & encryption", map[string]any{"engine": "sqlserver-ex", "storage_encrypted": true}, "sqlserver-ex does not support encryption"),
			Entry("sqlserver-se  & encryption", map[string]any{"engine": "sqlserver-se", "storage_encrypted": true}, ""),
			Entry("sqlserver-ee  & encryption", map[string]any{"engine": "sqlserver-ee", "storage_encrypted": true}, ""),
			Entry("sqlserver-web & encryption", map[string]any{"engine": "sqlserver-web", "storage_encrypted": true}, ""),
			Entry("sqlserver-ex  ! encryption", map[string]any{"engine": "sqlserver-ex", "storage_encrypted": false}, ""),
			Entry("sqlserver-se  ! encryption", map[string]any{"engine": "sqlserver-se", "storage_encrypted": false}, ""),
			Entry("sqlserver-ee  ! encryption", map[string]any{"engine": "sqlserver-ee", "storage_encrypted": false}, ""),
			Entry("sqlserver-web ! encryption", map[string]any{"engine": "sqlserver-web", "storage_encrypted": false}, ""),
		)
	})

	Context("db parameter group", func() {
		When("with default values", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars))
			})

			It("should create a parameter group", func() {
				Expect(ResourceCreationForType(plan, "aws_db_parameter_group")).To(HaveLen(1))
				Expect(AfterValuesForType(plan, "aws_db_parameter_group")).To(
					MatchKeys(IgnoreExtras, Keys{
						"name_prefix": ContainSubstring("rds-mssql-csb-mssql-test"),
						"family":      Equal("sqlserver-ee-15.0"),
						"parameter": ConsistOf(
							MatchKeys(IgnoreExtras, Keys{
								"name":         Equal("contained database authentication"),
								"value":        Equal("1"),
								"apply_method": Equal("immediate"),
							}),
						)},
					))
			})
		})
	})

	Context("maintenance_window", func() {
		When("no window is set", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{}))
			})
			It("should not be passed", func() {
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(Not(HaveKey("maintenance_window")))
			})
		})

		When("maintainance window specified with all values", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
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

	Context("auto_minor_version_upgrade", func() {
		When("is enabled and a not major version is selected", func() {
			It("should complain about postcondition", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"auto_minor_version_upgrade": true,
					"mssql_version":              "15.00.4236.7.v1",
				}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`Error: Resource postcondition failed`))
				Expect(msgs).To(ContainSubstring(`A Major engine version should be specified when auto_minor_version_upgrade is enabled. Expected engine version: 15.00 - got: 15.00.4236.7.v1`))
			})
		})

		When("is enabled and a major version is selected", func() {
			It("should not complain about postcondition", func() {
				plan := ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"auto_minor_version_upgrade": true,
					"mssql_version":              "15.00",
				}))

				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"auto_minor_version_upgrade": BeTrue(),
						"engine_version":             Equal("15.00"),
					}))
			})
		})

		When("is disabled and a major version is selected", func() {
			It("should not complain about postcondition", func() {
				plan := ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"auto_minor_version_upgrade": false,
					"mssql_version":              "15.00",
				}))

				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"auto_minor_version_upgrade": BeFalse(),
						"engine_version":             Equal("15.00"),
					}))
			})
		})

		When("is disabled and a minor version is selected", func() {
			It("should not complain about postcondition and create the instance", func() {
				plan := ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"auto_minor_version_upgrade": false,
					"mssql_version":              "15.00.4236.7.v1",
				}))

				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"auto_minor_version_upgrade": BeFalse(),
						"engine_version":             Equal("15.00.4236.7.v1"),
					}))
			})
		})
	})

	Context("performance insights", func() {
		When("performance insights is enabled", func() {
			It("works as expected", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"performance_insights_enabled":          true,
					"performance_insights_kms_key_id":       "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa",
					"performance_insights_retention_period": 93,
				}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"performance_insights_enabled":          BeTrue(),
						"performance_insights_kms_key_id":       Equal("arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
						"performance_insights_retention_period": BeNumerically("==", 93),
					}),
				)
			})
		})

		When("performance insights is disabled", func() {
			It("causes kms_key_id and retention_period to be ignored", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"performance_insights_enabled":          false,
					"performance_insights_kms_key_id":       "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa",
					"performance_insights_retention_period": 93,
				}))
				Expect(maps.Keys(AfterValuesForType(plan, "aws_db_instance").(map[string]any))).To(
					Not(ContainElements(
						"performance_insights_enabled",
						"performance_insights_kms_key_id",
						"performance_insights_retention_period",
					)),
				)
			})
		})

		When("a kms key id with invalid format is passed", func() {
			It("refuses to create the aws_db_instance", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"performance_insights_kms_key_id": "an-invalid-kms-key-id",
				}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`(an-invalid-kms-key-id) is an invalid ARN: arn: invalid prefix`))
			})
		})

		When("an invalid retention period is passed", func() {
			It("doesn't detect any errors and plan finishes succesfully", func() {
				// invalid values for the retention_period are handled by the IAAS not Terraform
				plan := ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"performance_insights_enabled":          true,
					"performance_insights_retention_period": 13,
				}))

				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"performance_insights_enabled":          BeTrue(),
						"performance_insights_retention_period": BeNumerically("==", 13),
					}),
				)
			})
		})
	})

	Context("multi az", func() {
		When("some invalid combinations are passed", func() {
			It("it doesn't fail during the plan stage, only during apply", func() {
				overrideIncompatibleDefaults := map[string]any{"storage_encrypted": false}

				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, overrideIncompatibleDefaults, requiredVars, map[string]any{
					"multi_az": true,
					"engine":   "sqlserver-ex",
				}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(
					MatchKeys(IgnoreExtras, Keys{
						"multi_az": BeTrue(),
						"engine":   Equal("sqlserver-ex"),
					}),
				)
			})
		})

		When("no custom security_group_ids passed and multi_az is enabled", func() {
			It("should create several security_group_rules", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"rds_vpc_security_group_ids": "",
					"multi_az":                   true,
				}))
				Expect(ResourceChangesNames(plan)).To(ConsistOf(
					"db_instance",
					"db_parameter_group",
					"rds-private-subnet",
					"rds-sg",
					"mssql_multiaz_tcp_egress",
					"mssql_multiaz_tcp_ingress",
					"mssql_multiaz_udp_egress",
					"mssql_multiaz_udp_ingress",
					"rds_inbound_access",
					"password",
					"username",
				))
			})
		})

		When("no custom security_group_ids passed and multi_az is disabled", func() {
			It("should create several security_group_rules", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
					"rds_vpc_security_group_ids": "",
					"multi_az":                   false,
				}))
				Expect(ResourceChangesNames(plan)).To(ConsistOf(
					"db_instance",
					"db_parameter_group",
					"rds-private-subnet",
					"rds-sg",
					"rds_inbound_access",
					"password",
					"username",
				))
			})
		})

		/*
			// Can't implement this test because existing validations reject dummy security_group_ids
			When("some security_group_ids passed and multi_az is enabled", func() {
				It("should not create any security_group_rules", func() {
					plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{
						"rds_vpc_security_group_ids": "id1,id2,id3",
						"multi_az": true,
					}))
					Expect(ResourceChangesNames(plan)).To(ConsistOf(
						"db_instance",
						"db_parameter_group",
						"rds-private-subnet",
						"rds-sg",
						"rds_inbound_access",
						"password",
						"username",
					))
				})
			})
		*/
	})
})
