package withoutcredhub_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", func() {
	It("can be accessed by an app", func() {
		env := apps.EnvVar{Name: "CH_CRED_HUB_URL", Value: ""}
		broker := brokers.Create(
			brokers.WithPrefix("csb-storage"),
			brokers.WithEnv(env),
		)
		defer broker.Delete()

		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-s3-bucket",
			"private",
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
		got := app.GET(filename)
		Expect(got).To(Equal(fileContent))

		By("deleting the file from bucket")
		app.DELETE(filename)
	})
})
