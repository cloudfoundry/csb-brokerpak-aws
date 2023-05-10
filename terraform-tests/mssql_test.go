package terraformtests

import (
	"path"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		"db_name":               "vsbdb",
		"labels":                map[string]string{"label1": "value1"},
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "mssql/provision")
		Init(terraformProvisionDir)
	})

	Context("with Default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
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
				"engine_version":       Equal("15.00.4236.7.v1"),
				"identifier":           Equal("csb-mssql-test"),
				"instance_class":       Equal("db.m6i.xlarge"),
				"tags":                 HaveKeyWithValue("label1", "value1"),
				"db_subnet_group_name": Equal("csb-mssql-test-p-sn"),
				"apply_immediately":    BeTrue(),
				"skip_final_snapshot":  BeTrue(),
				"license_model":        Equal("license-included"),
			}))
		})
	})
})
