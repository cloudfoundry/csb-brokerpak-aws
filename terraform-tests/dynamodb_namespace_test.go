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
	)

	defaultVars := map[string]any{
		"region": "fake-region",
		"prefix": "csb-fake-5368-489c-9f18-b53140316fb2",
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "dynamodb-namespace/provision")
		Init(terraformProvisionDir)
	})

	Context("default provisioning", func() {
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
