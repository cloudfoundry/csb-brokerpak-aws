package acceptance_tests_test

import (
	"encoding/json"
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

var _ = Describe("Aurora MySQL", Label("aurora-mysql"), func() {
	It("can be accessed by an app", Label("JDBC-m"), func() {
		var (
			userIn, userOut jdbcapp.AppResponseUser
			sslInfo         jdbcapp.MySQLSSLInfo
		)

		By("creating a service instance")

		params := map[string]any{
			"cluster_instances": 2,
			"instance_class":    "db.r5.large",
			"engine_version":    "8.0.mysql_aurora.3.02.2",
		}

		serviceInstance := services.CreateInstance(
			"csb-aws-aurora-mysql",
			services.WithPlan("default"),
			services.WithParameters(params))
		defer serviceInstance.Delete()

		By("pushing the unstarted apps")
		manifest := jdbcapp.ManifestFor(jdbcapp.MySQL)
		appWriter := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(manifest))
		appReader := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(manifest))
		defer apps.Delete(appWriter, appReader)

		By("binding the the writer app")
		binding := serviceInstance.Bind(appWriter)

		By("starting the writer app")
		apps.Start(appWriter)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("creating an entry using the writer app")
		value := random.Hexadecimal()
		response := appWriter.POST("", "?name=%s", value)

		responseBody, err := io.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())

		err = json.Unmarshal(responseBody, &userIn)
		Expect(err).NotTo(HaveOccurred())

		By("binding the reader app to the reader endpoint")
		serviceInstance.Bind(appReader, services.WithBindParameters(map[string]any{"reader_endpoint": true}))

		By("starting the reader app")
		apps.Start(appReader)

		By("getting the entry using the reader app")
		got := appReader.GET("%d", userIn.ID)

		err = json.Unmarshal([]byte(got), &userOut)
		Expect(err).NotTo(HaveOccurred())
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		By("verifying the DB connection utilises TLS")
		got = appWriter.GET("mysql-ssl")
		err = json.Unmarshal([]byte(got), &sslInfo)
		Expect(err).NotTo(HaveOccurred())

		Expect(strings.ToLower(sslInfo.VariableName)).To(Equal("ssl_cipher"))
		Expect(sslInfo.Value).NotTo(BeEmpty())

		By("deleting the entry using the writer app")
		appWriter.DELETE("%d", userIn.ID)

		By("pushing and binding an app for verifying non-TLS connection attempts")
		golangApp := apps.Push(apps.WithApp(apps.MySQL))
		serviceInstance.Bind(golangApp)
		apps.Start(golangApp)

		By("verifying interactions with TLS enabled")
		key, value := "key", "value"
		golangApp.PUT(value, key)
		got = golangApp.GET(key)
		Expect(got).To(Equal(value))

		By("verifying that non-TLS connections should fail")
		response = golangApp.GetRawResponse("%s?tls=false", key)
		defer response.Body.Close()
		Expect(response.StatusCode).To(Equal(http.StatusInternalServerError), "force TLS is enabled by default")
		b, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in TLS failure")
		Expect(string(b)).To(ContainSubstring("error connecting to database: failed to verify the connection"), "force TLS is enabled by default")
		Expect(string(b)).To(ContainSubstring("Error 1045 (28000): Access denied for user"), "postgresql client cannot connect to the postgres server due to invalid TLS")
	})
})
