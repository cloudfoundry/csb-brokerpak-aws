package upgrade_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeS3Test", Label("upgrade", "s3"), func() {
	Context("When upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-upgrade"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-aws-s3-bucket",
				"private",
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.S3))
			appTwo := apps.Push(apps.WithApp(apps.S3))
			defer apps.Delete(appOne, appTwo)

			By("binding the apps to the s3 service instance")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)
			apps.Start(appOne, appTwo)

			By("uploading a blob using the first app")
			blobNameOne := random.Hexadecimal()
			blobDataOne := random.Hexadecimal()
			appOne.PUT(blobDataOne, blobNameOne)

			By("downloading the blob using the second app")
			got := appTwo.GET(blobNameOne)
			Expect(got).To(Equal(blobDataOne))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading the service instance")
			serviceInstance.Upgrade()

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne)
			Expect(got).To(Equal(blobDataOne))

			By("updating the service instance")
			serviceInstance.Update("-c", `{"pab_block_public_policy": true}`)

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne)
			Expect(got).To(Equal(blobDataOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("binding the app to the instance again")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne)
			Expect(got).To(Equal(blobDataOne))

			By("checking that data can still be written and read")
			blobNameTwo := random.Hexadecimal()
			blobDataTwo := random.Hexadecimal()
			appOne.PUT(blobDataTwo, blobNameTwo)
			got = appTwo.GET(blobNameTwo)
			Expect(got).To(Equal(blobDataTwo))

			appOne.DELETE(blobNameOne)
			appOne.DELETE(blobNameTwo)
		})
	})
})
