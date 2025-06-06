package acceptance_tests_test

import (
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/jdbcapp"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("Aurora PostgreSQL", Label("aurora-postgresql"), func() {
	It("it works with the oldest supported version of Postgres", Label("JDBC-p", "PostgresOldest"), func() {
		testWithMultipleApps("13")
	})

	It("works with the latest supported version of Postgres", Label("JDBC-p", "PostgresLatestAGC"), func() {
		testWithMultipleApps("17")
	})

	It("works with the latest supported version of Postgres", Label("JDBC-p", "PostgresLatest"), func() {
		testWithMultipleApps("17")
	})
})

func testWithMultipleApps(postgresVersion string) {
	var (
		userIn, userOut jdbcapp.AppResponseUser
		sslInfo         jdbcapp.PostgresSSLInfo
	)

	By("creating a service instance")
	params := map[string]any{
		"engine_version":    postgresVersion,
		"cluster_instances": 2,
		"instance_class":    "db.t3.medium",
	}
	serviceInstance := services.CreateInstance(
		"csb-aws-aurora-postgresql",
		services.WithPlan("default"),
		services.WithParameters(params),
	)
	defer serviceInstance.Delete()

	By("pushing the unstarted apps")
	manifest := jdbcapp.ManifestFor(jdbcapp.PostgreSQL)
	appWriter := apps.Push(apps.WithApp(apps.JDBCTestAppPostgres), apps.WithManifest(manifest))
	appReader := apps.Push(apps.WithApp(apps.JDBCTestAppPostgres), apps.WithManifest(manifest))
	defer apps.Delete(appWriter, appReader)

	By("binding the the writer app")
	binding := serviceInstance.Bind(appWriter)

	By("starting the writer app")
	apps.Start(appWriter)

	By("checking that the app environment has a credhub reference for credentials")
	Expect(binding.Credential()).To(matchers.HaveCredHubRef)

	By("creating an entry using the writer app")
	value := random.Hexadecimal()
	appWriter.POSTf("", "?name=%s", value).ParseInto(&userIn)

	By("binding the reader app to the reader endpoint")
	serviceInstance.Bind(appReader, services.WithBindParameters(map[string]any{"reader_endpoint": true}))

	By("starting the reader app")
	apps.Start(appReader)

	By("getting the entry using the reader app")
	appReader.GETf("%d", userIn.ID).ParseInto(&userOut)
	Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

	By("verifying the DB connection utilises TLS")
	appWriter.GET("postgres-ssl").ParseInto(&sslInfo)

	Expect(sslInfo.SSL).To(BeTrue())
	Expect(sslInfo.Cipher).NotTo(BeEmpty())

	By("deleting the entry using the writer app")
	appWriter.DELETEf("%d", userIn.ID)

	By("pushing and binding an app for verifying non-TLS connection attempts")
	golangApp := apps.Push(apps.WithApp(apps.PostgreSQL))
	serviceInstance.Bind(golangApp)
	apps.Start(golangApp)

	By("verifying interactions with TLS enabled")
	schema, key, value := "newschema", "key", "value"
	golangApp.PUT("", schema)
	golangApp.PUTf(value, "%s/%s", schema, key)
	got := golangApp.GETf("%s/%s", schema, key).String()
	Expect(got).To(Equal(value))

	By("verifying that non-TLS connections should fail")
	response := golangApp.GETResponsef("%s/%s?tls=disable", schema, key)
	defer response.Body.Close()
	Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError), "force TLS is enabled by default")
	b, err := io.ReadAll(response.Body)
	Expect(err).ToNot(HaveOccurred(), "error reading response body in TLS failure")
	Expect(string(b)).To(ContainSubstring("failed to connect to database"), "force TLS is enabled by default")
	Expect(string(b)).To(ContainSubstring("SQLSTATE 28000"), "postgresql client cannot connect to the postgres server due to invalid TLS")
}
