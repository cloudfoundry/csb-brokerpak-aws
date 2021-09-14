package s3_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := helpers.CreateService("csb-aws-s3-bucket", "private")
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.S3)
		appTwo := helpers.AppPushUnstarted(apps.S3)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the s3 service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(HaveKey("credhub-ref"))

		By("uploading a file using the first app")
		filename := helpers.RandomHex()
		fileContent := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
		appOne.PUT(fileContent, filename)

		By("downloading the file using the second app")
		got := appTwo.GET(filename)
		Expect(got).To(Equal(fileContent))

		By("deleting the file from bucket using the second app")
		appTwo.DELETE(filename)
	})
})
