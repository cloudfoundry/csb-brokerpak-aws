package s3_test

import (
	"acceptancetests/helpers"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"time"
)

var _ = Describe("S3", func() {
	var serviceInstanceName string

	BeforeEach(func() {
		serviceInstanceName = helpers.RandomName("s3")
		helpers.CreateService("csb-aws-s3-bucket", "private", serviceInstanceName)
	})

	AfterEach(func() {
		helpers.DeleteService(serviceInstanceName)
	})

	It("can be accessed by an app", func() {
		By("building the app")
		appDir := helpers.AppBuild("./s3app")
		defer os.RemoveAll(appDir)

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted("s3", appDir)
		appTwo := helpers.AppPushUnstarted("s3", appDir)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the s3 service instance")
		bindingName := helpers.Bind(appOne, serviceInstanceName)
		helpers.Bind(appTwo, serviceInstanceName)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		creds := helpers.GetBindingCredential(appOne, "csb-aws-s3-bucket", bindingName)
		Expect(creds).To(HaveKey("credhub-ref"))

		By("uploading a file using the first app")
		filename := helpers.RandomString()
		fileContent := []byte(fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String()))
		helpers.HTTPPostFile(fmt.Sprintf("http://%s.%s/%s", appOne, helpers.DefaultSharedDomain(), filename), fileContent)

		By("downloading the file using the second app")
		got := helpers.HTTPGet(fmt.Sprintf("http://%s.%s/%s", appTwo, helpers.DefaultSharedDomain(), filename))
		Expect(got).To(Equal(string(fileContent)))

		By("deleting the file from bucket using the second app")
		helpers.HTTPDelete(fmt.Sprintf("http://%s.%s/%s", appTwo, helpers.DefaultSharedDomain(), filename))
	})
})
