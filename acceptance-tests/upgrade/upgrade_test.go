package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeTest", func() {
	Context("When upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := helpers.PushAndStartBroker(brokerName, releasedBuildDir)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := helpers.CreateServiceInBroker("csb-aws-s3-bucket", "private", brokerName)
			defer serviceInstance.Delete()

			By("pushing the unstarted app")
			testApp := helpers.AppPushUnstarted(apps.S3)
			defer helpers.AppDelete(testApp)

			By("binding the app to the s3 service instance")
			serviceInstance.Bind(testApp)

			By("starting the app")
			helpers.AppStart(testApp)

			By("uploading a file using the first app")
			filename := helpers.RandomHex()
			fileContent := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
			testApp.PUT(fileContent, filename)
			defer testApp.DELETE(filename)

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("downloading the file")
			got := testApp.GET(filename)
			Expect(got).To(Equal(fileContent))

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(testApp)

			By("binding the app to the instance again")
			serviceInstance.Bind(testApp)
			helpers.AppRestage(testApp)

			By("downloading the file")
			got = testApp.GET(filename)
			Expect(got).To(Equal(fileContent))
		})
	})
})
