package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Aurora PostgreSQL", Label("aurora-postgresql"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-aurora-postgresql", services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		appWriter := apps.Push(apps.WithApp(apps.PostgreSQL))
		appReader := apps.Push(apps.WithApp(apps.PostgreSQL))
		defer apps.Delete(appWriter, appReader)

		By("binding the the writer app")
		binding := serviceInstance.Bind(appWriter)

		By("binding the reader app to the reader endpoint")
		serviceInstance.Bind(appReader, services.WithBindParameters(map[string]any{"reader_endpoint": true}))

		By("starting the apps")
		apps.Start(appWriter, appReader)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("creating a schema using the writer app")
		schema := random.Name(random.WithMaxLength(8))
		appWriter.PUT("", schema)

		By("setting a key-value using the writer app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appWriter.PUT(value, "%s/%s", schema, key)

		By("getting the value using the reader app")
		got := appReader.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the writer app")
		appWriter.DELETE(schema)
	})
})
