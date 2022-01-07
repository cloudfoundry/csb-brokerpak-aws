package upgrade_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeTest", func() {
	Context("When upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(brokers.WithPrefix("csb-upgrade"), brokers.WithSourceDir(releasedBuildDir))
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-aws-s3-bucket",
				"private",
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app")
			testApp := apps.Push(apps.WithApp(apps.S3))
			defer apps.Delete(testApp)

			By("binding the app to the s3 service instance")
			binding := serviceInstance.Bind(testApp)

			By("starting the app")
			apps.Start(testApp)

			By("uploading a file using the first app")
			filename := random.Hexadecimal()
			fileContent := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
			testApp.PUT(fileContent, filename)
			defer testApp.DELETE(filename)

			By("pushing the development version of the broker")
			serviceBroker.UpdateSourceDir(developmentBuildDir)

			By("downloading the file")
			got := testApp.GET(filename)
			Expect(got).To(Equal(fileContent))

			By("deleting bindings created before the upgrade")
			binding.Unbind()

			By("binding the app to the instance again")
			serviceInstance.Bind(testApp)
			testApp.Restage()

			By("downloading the file")
			got = testApp.GET(filename)
			Expect(got).To(Equal(fileContent))
		})
	})
})
