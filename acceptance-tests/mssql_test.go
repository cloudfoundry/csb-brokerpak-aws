package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("MSSQL", Label("mssql"), func() {
	It("can be created", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-mssql",
			services.WithPlan("default"))
		defer serviceInstance.Delete()
	})
})
