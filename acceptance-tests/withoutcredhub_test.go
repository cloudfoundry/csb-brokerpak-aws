package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", Label("withoutcredhub"), func() {
	It("can be accessed by an app", func() {
		broker := brokers.Create(
			brokers.WithPrefix("csb-storage"),
			brokers.WithLatestEnv(),
			brokers.WithEnv(apps.EnvVar{Name: "CH_CRED_HUB_URL", Value: ""}),
		)
		defer broker.Delete()

		By("creating a service instance")
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
			services.WithBroker(broker),
			services.WithName(serviceName),
		)

		By("pushing the unstarted app")
		app := apps.Push(apps.WithApp(apps.S3))
		defer apps.Delete(app)

		By("binding the app to the storage service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("checking that the app environment does not a credhub reference for credentials")
		Expect(binding.Credential()).NotTo(matchers.HaveCredHubRef)

		By("uploading a file")
		filename := random.Hexadecimal()
		fileContent := fmt.Sprintf("This is a dummy file that will be uploaded the S3 at %s.", time.Now().String())
		app.PUT(fileContent, filename)

		By("downloading the file")
		got := app.GET(filename).String()
		Expect(got).To(Equal(fileContent))

		By("deleting the file from bucket")
		app.DELETE(filename)
	})
})
