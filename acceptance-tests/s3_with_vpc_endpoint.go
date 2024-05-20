package acceptance_tests

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 with allowed VPC", Label("VPCEndpointS3"), func() {
	It("should allow access from specified VPC", func() {
		allowedVPCID := os.Getenv("ALLOWED_VPC_ID")
		Expect(allowedVPCID).NotTo(BeEmpty(), "Environment variable ALLOWED_VPC_ID is not set")

		By("creating a service instance with allowed_aws_vpc_id set")
		serviceInstance := services.CreateInstance(
			"csb-aws-s3-bucket",
			services.WithPlan("default"),
			services.WithParameters(map[string]any{
				"allowed_aws_vpc_id": allowedVPCID,
			}),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		app := apps.Push(apps.WithApp(apps.S3))
		defer apps.Delete(app)

		By("binding the app to the s3 service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("uploading a file using the app")
		filename := random.Hexadecimal()
		fileContent := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
		app.PUT(fileContent, filename)

		By("downloading the file using the app")
		got := app.GET(filename).String()
		Expect(got).To(Equal(fileContent))

		By("unbinding the app from the service instance")
		binding.Unbind()

		By("updating the service instance with a fake allowed_aws_vpc_id")
		serviceInstance.Update(services.WithParameters(map[string]any{"allowed_aws_vpc_id": "vpc-12345678"}))

		By("binding the app to the s3 service instance to create a policy with a non-expected VPC")
		binding = serviceInstance.Bind(app)
		app.Restage()

		By("checking the app cannot access the bucket")
		httpResponse := app.GETResponse(filename)
		defer httpResponse.Body.Close()

		Expect(httpResponse).To(
			HaveHTTPStatus(
				http.StatusFailedDependency),
			"the connection is not possible from a VPC that is not allowed",
		)
		b, err := io.ReadAll(httpResponse.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in S3 API call")
		Expect(string(b)).To(ContainSubstring("api error AccessDenied: Access Denied"), "access denied due to the policy restriction")

		By("updating the service instance with a fake allowed_aws_vpc_id")
		serviceInstance.Update(services.WithParameters(map[string]any{"allowed_aws_vpc_id": allowedVPCID}))

		By("unbinding the app from the service instance to regenerate the binding again and change the policy")
		binding.Unbind()
		binding = serviceInstance.Bind(app)
		defer binding.Unbind()
		app.Restage()

		By("deleting the file from bucket using the app")
		app.DELETE(filename)

		By("verifying that the file no longer exists")
		httpResponse = app.GETResponse(filename)
		defer httpResponse.Body.Close()
		Expect(httpResponse).To(
			HaveHTTPStatus(
				http.StatusFailedDependency),
			"there is not file to be retrieved",
		)
		b, err = io.ReadAll(httpResponse.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in S3 API call after deletion")
		Expect(string(b)).To(ContainSubstring("operation error S3: GetObject, https response error StatusCode: 404"), "file does not exist in the bucket")
	})
})
