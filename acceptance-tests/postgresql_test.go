package acceptance_tests_test

import (
	"encoding/json"
	"io"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/jdbcapp"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("PostgreSQL", Label("postgresql"), func() {
	It("can be accessed by an app", Label("JDBC"), func() {
		var (
			userIn, userOut jdbcapp.AppResponseUser
			sslInfo         jdbcapp.PostgresSSLInfo
		)

		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-postgresql", services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		testExecutable, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())

		testPath := path.Dir(testExecutable)
		appManifest := path.Join(testPath, "apps", "jdbctestapp", "manifest.yml")
		appOne := apps.Push(apps.WithApp(apps.JDBCTestAppPostgres), apps.WithManifest(appManifest))
		appTwo := apps.Push(apps.WithApp(apps.JDBCTestAppPostgres), apps.WithManifest(appManifest))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(appOne)

		By("starting the first app")
		apps.Start(appOne)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("creating an entry using the first app")
		value := random.Hexadecimal()
		response := appOne.POST("", "?name=%s", value)
		responseBody, err := io.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())
		err = json.Unmarshal(responseBody, &userIn)
		Expect(err).NotTo(HaveOccurred())

		By("binding and starting the second app")
		serviceInstance.Bind(appTwo)
		apps.Start(appTwo)

		By("getting the entry using the second app")
		got := appTwo.GET("%d", userIn.ID)

		err = json.Unmarshal([]byte(got), &userOut)
		Expect(err).NotTo(HaveOccurred())
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		By("verifying the DB connection utilises TLS")
		got = appOne.GET("postgres-ssl")
		err = json.Unmarshal([]byte(got), &sslInfo)
		Expect(err).NotTo(HaveOccurred())

		Expect(sslInfo.SSL).To(BeTrue())
		Expect(sslInfo.Cipher).NotTo(BeEmpty())
		Expect(sslInfo.Bits).To(BeNumerically(">=", 256))

		By("deleting the entry using the first app")
		appOne.DELETE("%d", userIn.ID)
	})
})
