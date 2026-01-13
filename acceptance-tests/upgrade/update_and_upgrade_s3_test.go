package upgrade_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/plans"
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
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceOffering := "csb-aws-s3-bucket"
			servicePlan := "default"
			serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
			// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
			// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
			// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
			// (which can only have 5 instances at any time) to prevent subsequent test failures.
			defer services.Delete(serviceName)
			serviceInstance := services.CreateInstance(
				serviceOffering,
				services.WithPlan(servicePlan),
				services.WithBroker(serviceBroker),
				services.WithName(serviceName),
			)

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.S3))
			appTwo := apps.Push(apps.WithApp(apps.S3))
			defer apps.Delete(appOne, appTwo)

			By("binding the apps to the s3 service instance")
			// Using WithBindParameters to ensure bind request details are stored in the database,
			// which exercises the upgrade path for the bind_resource migration (PR #1341)
			bindingOne := serviceInstance.Bind(appOne, services.WithBindParameters(`{}`))
			bindingTwo := serviceInstance.Bind(appTwo, services.WithBindParameters(`{}`))
			apps.Start(appOne, appTwo)

			By("uploading a blob using the first app")
			blobNameOne := random.Hexadecimal()
			blobDataOne := random.Hexadecimal()
			appOne.PUT(blobDataOne, blobNameOne)

			By("downloading the blob using the second app")
			got := appTwo.GET(blobNameOne).String()
			Expect(got).To(Equal(blobDataOne))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable(servicePlan, serviceOffering, serviceBroker.Name))

			By("upgrading the service instance")
			serviceInstance.Upgrade()

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne).String()
			Expect(got).To(Equal(blobDataOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("binding the app to the instance again")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("updating the service instance")
			serviceInstance.Update(services.WithParameters(`{}`))

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne).String()
			Expect(got).To(Equal(blobDataOne))

			By("deleting bindings created before the update")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("binding the app to the instance again")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking that previously written data is accessible")
			got = appTwo.GET(blobNameOne).String()
			Expect(got).To(Equal(blobDataOne))

			By("checking that data can still be written and read")
			blobNameTwo := random.Hexadecimal()
			blobDataTwo := random.Hexadecimal()
			appOne.PUT(blobDataTwo, blobNameTwo)
			got = appTwo.GET(blobNameTwo).String()
			Expect(got).To(Equal(blobDataTwo))

			appOne.DELETE(blobNameOne)
			appOne.DELETE(blobNameTwo)
		})
	})
})
