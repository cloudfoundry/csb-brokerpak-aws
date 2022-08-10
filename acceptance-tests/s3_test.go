package acceptance_tests_test

import (
	"fmt"
	"time"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3", Label("s3"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-s3-bucket", services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.S3))
		appTwo := apps.Push(apps.WithApp(apps.S3))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the s3 service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(HaveKey("credhub-ref"))

		By("uploading a file using the first app")
		filename := random.Hexadecimal()
		fileContent := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
		appOne.PUT(fileContent, filename)

		By("downloading the file using the second app")
		got := appTwo.GET(filename)
		Expect(got).To(Equal(fileContent))

		By("deleting the file from bucket using the second app")
		appTwo.DELETE(filename)
	})
})
