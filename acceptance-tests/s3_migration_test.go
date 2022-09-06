package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/dms"
	"fmt"
	"time"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 Migration", Label("s3-migration"), func() {
	It("can migrate instance from the previous legacy broker to the CSB", func() {
		By("creating an S3 bucket using the legacy broker")
		legacyServiceInstance := services.CreateInstance(
			"aws-s3",
			services.WithPlan("standard"),
			services.WithBrokerName("aws-services-broker"),
		)

		By("pushing an unstarted app")
		appOne := apps.Push(apps.WithApp(apps.S3))
		defer apps.Delete(appOne)

		By("binding the apps to the legacy service instance")
		legacyBinding := legacyServiceInstance.Bind(appOne)

		By("starting the app")
		apps.Start(appOne)

		By("uploading files using the first app")
		filename1 := random.Hexadecimal()
		fileContent1 := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
		appOne.PUT(fileContent1, filename1)

		filename2 := random.Hexadecimal()
		fileContent2 := fmt.Sprintf("This is another dummy file that will be uploaded the S3 at %s.", time.Now().String())
		appOne.PUT(fileContent2, filename2)

		By("creating an S3 bucket using the CSB broker")
		targetServiceInstance := services.CreateInstance(
			"csb-aws-s3-bucket",
			services.WithPlan("default"),
			services.WithDefaultBroker(),
		)
		defer targetServiceInstance.Delete()

		By("syncing up the legacy S3 to the new S3")
		sourceBucketLocation := "s3://cf-" + legacyServiceInstance.GUID()
		targetBucketLocation := "s3://csb-" + targetServiceInstance.GUID()
		result := dms.AWS("s3", "sync", sourceBucketLocation, targetBucketLocation)
		Expect(result).To(SatisfyAll(
			ContainSubstring(fmt.Sprintf("copy: %s/%s to %s/%s", sourceBucketLocation, filename1, targetBucketLocation, filename1)),
			ContainSubstring(fmt.Sprintf("copy: %s/%s to %s/%s", sourceBucketLocation, filename2, targetBucketLocation, filename2)),
		))

		By("switching the app data source from the legacy S3 to the new S3")
		legacyBinding.Unbind()
		targetServiceInstance.Bind(appOne)
		apps.Restart(appOne)

		By("checking existing data can be accessed using the new S3 instance")
		Expect(appOne.GET(filename1)).To(Equal(fileContent1))
		Expect(appOne.GET(filename2)).To(Equal(fileContent2))

		By("deleting the file from bucket using the second app")
		appOne.DELETE(filename1)
		appOne.DELETE(filename2)
	})
})
