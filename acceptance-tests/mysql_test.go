package acceptance_tests_test

import (
	"io"
	"net/http"
	"strings"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/jdbcapp"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MySQL", Label("mysql"), func() {
	It("can be accessed by an app", Label("JDBC-m"), func() {
		var (
			userIn, userOut jdbcapp.AppResponseUser
			sslInfo         jdbcapp.MySQLSSLInfo
		)

		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-mysql",
			services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		manifest := jdbcapp.ManifestFor(jdbcapp.MySQL)

		appOne := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(manifest))
		appTwo := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(manifest))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("creating an entry using the first app")
		value := random.Hexadecimal()
		appOne.POST("", "?name=%s", value).ParseInto(&userIn)

		By("binding and starting the second app")
		serviceInstance.Bind(appTwo)
		apps.Start(appTwo)

		By("getting the entry using the second app")
		appTwo.GET("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		By("verifying the DB connection utilises TLS")
		appOne.GET("mysql-ssl").ParseInto(&sslInfo)
		Expect(strings.ToLower(sslInfo.VariableName)).To(Equal("ssl_cipher"))
		Expect(sslInfo.Value).NotTo(BeEmpty())

		By("deleting the entry using the first app")
		appOne.DELETE("%d", userIn.ID)

		By("pushing and binding an app for verifying non-TLS connection attempts")
		golangApp := apps.Push(apps.WithApp(apps.MySQL))
		serviceInstance.Bind(golangApp)
		apps.Start(golangApp)

		By("verifying interactions with TLS enabled")
		const key = "key"
		value = "value"
		golangApp.PUT(value, key)
		got := golangApp.GET(key).String()
		Expect(got).To(Equal(value))

		By("verifying that non-TLS connections should fail")
		response := golangApp.GETResponse("%s?tls=false", key)
		defer response.Body.Close()
		Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError), "force TLS is enabled by default")
		b, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in TLS failure")
		Expect(string(b)).To(ContainSubstring("error connecting to database: failed to verify the connection"), "force TLS is enabled by default")
		Expect(string(b)).To(ContainSubstring("Error 1045 (28000): Access denied for user"), "postgresql client cannot connect to the postgres server due to invalid TLS")
	})
})
