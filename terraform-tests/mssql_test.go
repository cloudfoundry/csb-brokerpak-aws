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

	// Most default values are automatically loaded from /terraform/mssql/provision/variables.tf
	// Having them there poses the same amount of duplication than having them in this file.
	// However, having them there comes with some additional benefits.
	defaultVars := map[string]any{
		"aws_access_key_id":          awsAccessKeyID,
		"aws_secret_access_key":      awsSecretAccessKey,
		"aws_vpc_id":                 "",
		"db_name":                    "vsbdb",
		"instance_name":              "csb-mssql-test",
		"instance_class":             "",
		"labels":                     map[string]string{"label1": "value1"},
		"region":                     "us-west-2",
		"rds_subnet_group":           "",
		"rds_vpc_security_group_ids": "",
	}

	requiredVars := map[string]any{
		"engine":        "sqlserver-ee",
		"mssql_version": "",
		"storage_gb":    20,
	}

	validVPC := awsVPCID

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "mssql/provision")
		Init(terraformProvisionDir)
	})

	Context("Creating an Instance", func() {
		When("only passing the default values", func() {
			It("should complain about missing required values", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars))
				Expect(session.ExitCode()).NotTo(Equal(0))
				msgs := string(session.Out.Contents())
				Expect(msgs).To(ContainSubstring(`The root module input variable \"mssql_version\" is not set, and has no default value.`))
				Expect(msgs).To(ContainSubstring(`The root module input variable \"engine\" is not set, and has no default value.`))
				Expect(msgs).To(ContainSubstring(`The root module input variable \"storage_gb\" is not set, and has no default value.`))
			})
		})

		When("with all required fields satisfied", func() {
			It("should create a db instance with the right values", func() {
				plan := ShowPlan(terraformProvisionDir, buildVars(defaultVars, requiredVars))
				Expect(plan.ResourceChanges).To(HaveLen(6))

				Expect(ResourceChangesTypes(plan)).To(ConsistOf(
					"aws_db_instance",
					"random_password",
					"random_string",
					"aws_security_group_rule",
					"aws_db_subnet_group",
					"aws_security_group",
				))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{
					"engine":               Equal("sqlserver-ee"),
					"identifier":           Equal("csb-mssql-test"),
					"instance_class":       Equal(""),
					"tags":                 HaveKeyWithValue("label1", "value1"),
					"db_subnet_group_name": Equal("csb-mssql-test-p-sn"),
					"apply_immediately":    BeTrue(),
					"skip_final_snapshot":  BeTrue(),
					"license_model":        Equal("license-included"),
				}))
			})
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
})
