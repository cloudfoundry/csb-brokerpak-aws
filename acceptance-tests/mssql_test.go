package acceptance_tests_test

import (
	"io"
	"net/http"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/jdbcapp"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL", Label("mssql"), func() {
	It("can be created", func() {
		By("creating a service instance")
		params := map[string]any{
			// The backups add quite some time to acceptance test.
			// Every time we test the service using the cf CLI we do it with backups enabled as it is the
			// default option. The idea is to speed up the execution of the test.
			"backup_retention_period": 0,
			"require_ssl":             true,
		}

		serviceInstance := services.CreateInstance(
			"csb-aws-mssql",
			services.WithPlan("default"),
			services.WithParameters(params),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.MSSQL))
		appTwo := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("creating a schema using the first app")
		schema := random.Name(random.WithMaxLength(10))
		appOne.PUT("", schema)

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appOne.PUT(value, "%s/%s", schema, key)

		By("getting the value using the second app")
		got := appTwo.GET("%s/%s", schema, key).String()
		Expect(got).To(Equal(value))

		By("verifying that non-TLS connections should fail")
		response := appTwo.GETResponse("%s/%s?tls=disable", schema, key)
		defer response.Body.Close()
		Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError), "force TLS is enabled by default")
		b, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in TLS failure")
		Expect(string(b)).To(ContainSubstring("TLS Handshake failed: cannot read handshake packet: EOF"), "force TLS is enabled by default")

		By("deleting binding one the binding two keeps reading the value")
		binding.Unbind()
		got = appTwo.GET("%s/%s", schema, key).String()
		Expect(got).To(Equal(value))

		By("dropping the schema using the second app")
		appTwo.DELETE(schema)
	})

	It("can be accessed by a JAVA app using the JDBC URL", Label("JDBC-mssql"), func() {
		var (
			userIn  jdbcapp.AppResponseUser
			userOut jdbcapp.AppResponseUser
		)

		By("creating a service instance")
		params := map[string]any{
			// The backups add quite some time to acceptance test.
			// Every time we test the service using the cf CLI we do it with backups enabled as it is the
			// default option. The idea is to speed up the execution of the test.
			"backup_retention_period": 0,
			"require_ssl":             true,
		}

		serviceInstance := services.CreateInstance(
			"csb-aws-mssql",
			services.WithPlan("default"),
			services.WithParameters(params),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		manifest := jdbcapp.ManifestFor(jdbcapp.SQLServer)
		appWriter := apps.Push(apps.WithApp(apps.JDBCTestAppSQLServer), apps.WithManifest(manifest))
		appReader := apps.Push(apps.WithApp(apps.JDBCTestAppSQLServer), apps.WithManifest(manifest))
		defer apps.Delete(appWriter, appReader)

		By("binding the the writer app")
		serviceInstance.Bind(appWriter)

		By("starting the writer app")
		apps.Start(appWriter)

		By("creating an entry using the writer app")
		value := random.Hexadecimal()
		appWriter.POST("", "?name=%s", value).ParseInto(&userIn)

		By("binding the reader app")
		serviceInstance.Bind(appReader)

		By("starting the reader app")
		apps.Start(appReader)

		By("getting the entry using the reader app")
		appReader.GET("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		// This step is not necessary, added for the purpose of serving as documentation
		By("verifying the DB connection utilises TLS")
		httpResponse := appWriter.GETResponse("sqlserver-ssl")
		defer httpResponse.Body.Close()
		Expect(httpResponse.StatusCode).To(BeNumerically("==", http.StatusInternalServerError), "it can be run only by administrators with the VIEW SERVER STATE privilege")
	})

	It("can't be destroyed if `deletion_protection: true`", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-mssql",
			services.WithPlan("default"),
			services.WithParameters(map[string]any{
				"deletion_protection": true,
			}),
		)
		err := InterceptGomegaFailure(func() { serviceInstance.Delete() })
		Expect(err).To(HaveOccurred())

		serviceInstance.Update(
			services.WithParameters(map[string]any{
				"deletion_protection": false,
			}),
		)
		serviceInstance.Delete()
	})
})
