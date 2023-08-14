package terraformtests

import (
	"path"

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

		"aws_vpc_id":                 "",
		"rds_subnet_group":           "",
		"rds_vpc_security_group_ids": "",

		"storage_type": "io1",
		"iops":         3000,
	}

	requiredVars := map[string]any{
		"engine":        "sqlserver-ee",
		"mssql_version": "some-engine-version",
		"storage_gb":    20,

		"instance_class": "some-instance-class",
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
			Expect(plan.ResourceChanges).To(HaveLen(6))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_db_instance",
				"random_password",
				"random_string",
				"aws_security_group_rule",
				"aws_db_subnet_group",
				"aws_security_group",
			))
		})

		It("should create a db instance with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{
				"engine":               Equal("sqlserver-ee"),
				"engine_version":       Equal("some-engine-version"),
				"identifier":           Equal("csb-mssql-test"),
				"storage_encrypted":    BeTrue(),
				"instance_class":       Equal("some-instance-class"),
				"tags":                 HaveKeyWithValue("label1", "value1"),
				"db_subnet_group_name": Equal("csb-mssql-test-p-sn"),
				"apply_immediately":    BeTrue(),
				"skip_final_snapshot":  BeTrue(),
				"license_model":        Equal("license-included"),
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

	Context("properties that pass-through without validation", func() {
		When("mssql_version is passed", func() {
			It("it passes-through without validation", func() {
				// It would be possible to validate the engine version before apply using
				// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/rds_engine_version
				// It has some very interesting properties and use cases, such as:
				//  - valid_upgrade_targets - Set of engine versions that this database engine version can be upgraded to.
				//  - supports_read_replica - Indicates whether the database engine version supports read replicas.
				//  - supports_log_exports_to_cloudwatch - Indicates whether the engine version supports exporting the log types specified by
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars, map[string]any{"mssql_version": "ANY-VALUE-AT-ALL"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"engine_version": Equal("ANY-VALUE-AT-ALL")}))
			})
		})
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
					"aws_db_subnet_group",
					"aws_security_group",
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
})
