package terraformtests

import (
	"path"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		"aws_access_key_id":     awsAccessKeyID,
		"aws_secret_access_key": awsSecretAccessKey,
		"instance_name":         "csb-mssql-test",
		"labels":                map[string]string{"label1": "value1"},
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "mssql/provision")
		Init(terraformProvisionDir)
	})

	Context("with Default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, nil))
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

	Context("properties that pass-through without validation", func() {
		When("mssql_version is passed", func() {
			It("it passes-through without validation", func() {
				// It would be possible to validate the engine version before apply using
				// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/rds_engine_version
				// It has some very interesting properties and use cases, such as:
				//  - valid_upgrade_targets - Set of engine versions that this database engine version can be upgraded to.
				//  - supports_read_replica - Indicates whether the database engine version supports read replicas.
				//  - supports_log_exports_to_cloudwatch - Indicates whether the engine version supports exporting the log types specified by
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"mssql_version":"ANY-VALUE-AT-ALL"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"engine_version": Equal("ANY-VALUE-AT-ALL")}))
			})
		})
		When("instance_class is passed", func() {
			It("it passes-through without validation", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"instance_class":"ANY-VALUE-AT-ALL"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"instance_class": Equal("ANY-VALUE-AT-ALL")}))
			})
		})
	})
	Context("instance_name", func() {
		When("invalid instance_name is passed", func() {
			It("fails and returns a descriptive message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"instance_name":"THIS-ENGINE-DOESNT-EXIST"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session.Err).To(gbytes.Say("only lowercase alphanumeric characters, hyphens, underscores, periods, and spaces allowed in \"name\""))
				
			})
		})
		When("instance_name is passed", func() {
			It("it is used to populate several fields", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"instance_name":"some-better-value"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"identifier": Equal("some-better-value")}))
				Expect(AfterValuesForType(plan, "aws_security_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("some-better-value-sg")}))
				Expect(AfterValuesForType(plan, "aws_db_subnet_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("some-better-value-p-sn")}))
			})
		})
	})

	Context("engine", func() {
		When("no engine passed", func() {
			It("should use the default value", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, nil))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"engine": Equal("sqlserver-ee")}))
			})
		})

		When("valid engine passed", func() {
			It("should use the passed value", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"engine":"sqlserver-web"}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"engine": Equal("sqlserver-web")}))
			})
		})

		When("invalid engine passed", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"engine":"THIS-ENGINE-DOESNT-EXIST"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session.Err).To(gbytes.Say("Invalid value for variable"))
			})
		})

		When("valid rds engine but not sqlserver edition", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"engine":"mysql"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session.Err).To(gbytes.Say("Invalid value for variable"))
			})
		})
	})

	Context("storage_gb", func() {
		When("no storage_gb passed", func() {
			It("should use the default value", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, nil))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"allocated_storage": Equal(float64(20))}))
			})
		})

		When("valid storage_gb passed", func() {
			It("should use the passed value", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"storage_gb":60}))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"allocated_storage": Equal(float64(60))}))
			})
		})

		When("storage_gb is above the allowed maximum", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"storage_gb":4097}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session.Err).To(gbytes.Say("Invalid value for variable"))
			})
		})

		When("storage_gb is below the allowed minimum", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"storage_gb":19}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session.Err).To(gbytes.Say("Invalid value for variable"))
			})
		})
	})

	Context("db_name", func() {
		When("passed", func() {
			It("is not assigned to the db_instance", func() {
				// This behaviour is intentional. AWS RDS for SQLServer doesn't allow specifying a db_name.
				// The db_name will be passed to the binding's provider which will be responsible for creating the db.
				// https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"db_name":"ANY-VALUE-AT-ALL"}))
				Expect(AfterValuesForType(plan, "")).To(BeNil())
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras|IgnoreMissing, Keys{"db_name": BeNil()}))
			})
		})
	})

	Context("aws_vpc_id", func() {
		When("no vpc passed", func() {
			It("should succeed and use the default one", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, nil))
				Expect(session.ExitCode()).To(Equal(0))
			})
		})

		When("valid vpc passed", func() {
			It("should create the right resources", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"aws_vpc_id":awsVPCID}))

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

		When("invalid subnet group passed", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"aws_vpc_id":"THIS-VPC-DOESNT-EXIST"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session.Err).To(gbytes.Say("no matching EC2 VPC found"))
			})
		})
	})

	Context("subnet group", func() {
		When("no subnet group passed", func() {
			It("should create a new one", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, nil))
				Expect(ResourceCreationForType(plan, "aws_db_subnet_group")).To(HaveLen(1))
				Expect(AfterValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"db_subnet_group_name": Equal("csb-mssql-test-p-sn")}))
				Expect(AfterValuesForType(plan, "aws_db_subnet_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("csb-mssql-test-p-sn")}))
			})
		})

		When("invalid subnet group passed", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"rds_subnet_group":"THIS-SUBNET-GROUP-DOESNT-EXIST"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session.Err).To(gbytes.Say("no matching RDS DB Subnet Group found"))
			})
		})
	})

	Context("security groups", func() {
		When("no security group ids passed", func() {
			It("should create a new one", func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, nil))
				Expect(ResourceCreationForType(plan, "aws_security_group")).To(HaveLen(1))
				Expect(UnknownValuesForType(plan, "aws_db_instance")).To(MatchKeys(IgnoreExtras, Keys{"vpc_security_group_ids": BeTrue()}))
				Expect(AfterValuesForType(plan, "aws_security_group")).To(MatchKeys(IgnoreExtras, Keys{"name": Equal("csb-mssql-test-sg")}))
			})
		})

		When("invalid security group ids passed", func() {
			It("should fail and return a descriptive error message", func() {
				session, _ := FailPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"rds_vpc_security_group_ids":"THESE,SECURITY-GROUPS,DONT-EXIST"}))

				Expect(session.ExitCode()).NotTo(Equal(0))
				Expect(session.Err).To(gbytes.Say("Resource postcondition failed"))
			})
		})
	})
})
