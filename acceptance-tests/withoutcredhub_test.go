package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", Label("withoutcredhub"), func() {
	It("can be accessed by an app", func() {
		broker := brokers.Create(
			brokers.WithPrefix("csb-storage"),
			brokers.WithLatestEnv(),
			brokers.WithEnv(apps.EnvVar{Name: "CH_CRED_HUB_URL", Value: ""}),
		)
		defer broker.Delete()

		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-s3-bucket",
			services.WithPlan("default"),
			services.WithBroker(broker),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		app := apps.Push(apps.WithApp(apps.S3))
		defer apps.Delete(app)

		By("binding the app to the storage service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("checking that the app environment does not a credhub reference for credentials")
		Expect(binding.Credential()).NotTo(matchers.HaveCredHubRef)

		By("uploading a file")
		filename := random.Hexadecimal()
		fileContent := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
		app.PUT(fileContent, filename)

		By("downloading the file")
		got := app.GET(filename).String()
		Expect(got).To(Equal(fileContent))

		By("deleting the file from bucket")
		app.DELETE(filename)
	})
})
