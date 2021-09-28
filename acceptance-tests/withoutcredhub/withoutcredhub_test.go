package withoutcredhub_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := helpers.CreateService("csb-aws-s3-bucket", "private")
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		app := helpers.AppPushUnstarted(apps.S3)
		defer helpers.AppDelete(app)

		By("binding the app to the storage service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		helpers.AppStart(app)

		By("checking that the app environment does not a credhub reference for credentials")
		Expect(binding.Credential()).NotTo(helpers.HaveCredHubRef)

		By("uploading a file")
		filename := helpers.RandomHex()
		fileContent := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
		app.PUT(fileContent, filename)

		By("downloading the file")
		got := app.GET(filename)
		Expect(got).To(Equal(fileContent))

		By("deleting the file from bucket")
		app.DELETE(filename)
	})
})
