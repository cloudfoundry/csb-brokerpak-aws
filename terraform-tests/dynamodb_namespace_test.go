package terraformtests

import (
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "csbbrokerpakaws/terraform-tests/helpers"
)

var _ = Describe("dynamodb-namespace", Label("dynamodb-ns-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
		defaultVars           map[string]any
	)

	Describe("provisioning", func() {
		BeforeAll(func() {
			terraformProvisionDir = path.Join(workingDir, "dynamodb-namespace/provision")
			defaultVars = map[string]any{
				"region": "fake-region",
				"prefix": "csb-fake-5368-489c-9f18-b53140316fb2",
			}
			Init(terraformProvisionDir)
		})

		Context("default", func() {
			BeforeAll(func() {
				plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
			})

			It("should not create any resources", func() {
				Expect(plan.ResourceChanges).To(BeEmpty())
			})

			It("should pass through the parameters", func() {
				Expect(plan.OutputChanges).To(HaveKeyWithValue("region", BeAssignableToTypeOf(&tfjson.Change{})))
				Expect(plan.OutputChanges).To(HaveKeyWithValue("prefix", BeAssignableToTypeOf(&tfjson.Change{})))
				Expect(plan.OutputChanges["region"].After).To(Equal("fake-region"))
				Expect(plan.OutputChanges["prefix"].After).To(Equal("csb-fake-5368-489c-9f18-b53140316fb2"))
			})
		})
	})

	Describe("binding", func() {
		BeforeAll(func() {
			terraformProvisionDir = path.Join(workingDir, "dynamodb-namespace/bind")
			defaultVars = map[string]any{
				"user_name":             "fake-user-name",
				"prefix":                "csb-fake-5368-489c-9f18-b53140316fb2",
				"region":                "us-west-1",
				"aws_access_key_id":     awsAccessKeyID,
				"aws_secret_access_key": awsSecretAccessKey,
			}
			Init(terraformProvisionDir)
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))

			Expect(awsRegion).NotTo(BeEmpty(), "AWS must be provided in GSB_PROVISION_DEFAULTS")
		})

		It("should include the new user credentials", func() {
			Expect(plan.OutputChanges).To(HaveKeyWithValue("access_key_id", BeAssignableToTypeOf(&tfjson.Change{})))
			Expect(plan.OutputChanges).To(HaveKeyWithValue("secret_access_key", BeAssignableToTypeOf(&tfjson.Change{})))
		})
	})
})
