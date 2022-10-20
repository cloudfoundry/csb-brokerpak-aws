package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Aurora MySQL", Label("aurora-mysql"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-aurora-mysql",
			services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appWriter := apps.Push(apps.WithApp(apps.MySQL))
		appReader := apps.Push(apps.WithApp(apps.MySQL))
		defer apps.Delete(appWriter, appReader)

		By("binding the the writer")
		binding := serviceInstance.Bind(appWriter)

		By("binding the reader app as 'readonly'")
		serviceInstance.Bind(appReader, services.WithBindParameters(map[string]any{"reader_endpoint": true}))

		By("starting the apps")
		apps.Start(appWriter, appReader)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("setting and getting a key-value using the writer app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appWriter.PUT(value, key)
		got := appWriter.GET(key)
		Expect(got).To(Equal(value))

		By("getting the value using the reader app")
		got = appReader.GET(key)
		Expect(got).To(Equal(value))
	})
})
