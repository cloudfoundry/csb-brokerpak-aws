package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"csbbrokerpakaws/acceptance-tests/helpers/jdbcapp"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
	"fmt"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("MSSQL", Label("mssql"), func() {
	It("can be accessed by a JAVA app using the JDBC URL", Label("mssql-JDBC-tls"), func() {
		var (
			userIn  jdbcapp.AppResponseUser
			userOut jdbcapp.AppResponseUser
		)

		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-mssql", services.WithPlan("default"))
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
		appWriter.POSTf("", "?name=%s", value).ParseInto(&userIn)

		By("binding the reader app")
		serviceInstance.Bind(appReader)

		By("starting the reader app")
		apps.Start(appReader)

		By("getting the entry using the reader app")
		appReader.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		// This step is not necessary, added for the purpose of serving as documentation
		By("verifying the DB connection utilises TLS")
		httpResponse := appWriter.GETResponse("sqlserver-ssl")
		defer httpResponse.Body.Close()
		Expect(httpResponse.StatusCode).To(BeNumerically("==", http.StatusInternalServerError), "it can be run only by administrators with the VIEW SERVER STATE privilege")

		By("pushing and binding two apps for verifying non-TLS connection attempts and object reassignment")
		golangAppOne := apps.Push(apps.WithApp(apps.MSSQL))
		golangAppTwo := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(golangAppOne, golangAppTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(golangAppOne)
		serviceInstance.Bind(golangAppTwo)

		By("starting the apps")
		apps.Start(golangAppOne, golangAppTwo)
		By("creating a schema using the first app")
		schema := random.Name(random.WithMaxLength(10))
		golangAppOne.PUTf("", "%s?dbo=false", schema)

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value = random.Hexadecimal()
		golangAppOne.PUTf(value, "%s/%s", schema, key)

		By("verifying that non-TLS connections should fail")
		response := golangAppTwo.GETResponsef("%s/%s?tls=disable", schema, key)
		defer response.Body.Close()
		Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError), "force TLS is enabled by default")
		b, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in TLS failure")
		Expect(string(b)).To(ContainSubstring("TLS Handshake failed: cannot read handshake packet:"), "force TLS is enabled by default")

		By("deleting binding one the binding two keeps reading the value - object reassignment works")
		binding.Unbind()
		got := golangAppTwo.GETf("%s/%s", schema, key).String()
		Expect(got).To(Equal(value))

		By("dropping the schema using the second app")
		golangAppTwo.DELETE(schema)
	})

	It("can be accessed by a JAVA app using the JDBC URL when require ssl is disabled", Label("mssql-JDBC-notls"), func() {
		var (
			userIn  jdbcapp.AppResponseUser
			userOut jdbcapp.AppResponseUser
		)

		By("creating a service instance")
		params := map[string]any{
			"backup_retention_period": 0,
			"require_ssl":             false,
			"multi_az":                false,
		}

		serviceInstance := services.CreateInstance(
			"csb-aws-mssql",
			services.WithPlan("default"),
			services.WithParameters(params),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		manifest := jdbcapp.ManifestFor(jdbcapp.SQLServer)
		appWriter := apps.Push(apps.WithApp(apps.JDBCTestAppSQLServer), apps.WithManifest(manifest))
		defer apps.Delete(appWriter)

		By("binding the the app")
		serviceInstance.Bind(appWriter)

		By("starting the writer app")
		apps.Start(appWriter)

		By("creating an entry using the writer app")
		value := random.Hexadecimal()
		appWriter.POSTf("", "?name=%s", value).ParseInto(&userIn)

		By("getting the entry using the reader app")
		appWriter.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value))
	})

	It("can't be destroyed if `deletion_protection: true`", Label("mssql-deletion-protection"), func() {
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

	// As we introduce the 'use_managed_admin_password' feature, some users may wish to update existing DBs.
	// This is a tactical test that should exist for this changeover period and is not intended to be a forever test.
	// Due to limitations in Tofu/AWS provider/AWS the operation to switch fails first time, then succeeds on
	// a second attempt. That's not an ideal customer experience, and this test exists to ensure that what we
	// document works, and make us aware if the behavior changes.
	It("allows 'use_managed_admin_password' to be enabled", Label("managed-password"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-mssql", services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing an unstarted app")
		appManifest := jdbcapp.ManifestFor(jdbcapp.SQLServer)
		app := apps.Push(apps.WithApp(apps.JDBCTestAppSQLServer), apps.WithManifest(appManifest))
		defer apps.Delete(app)

		By("binding the app to the service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating an entry using the app")
		value := random.Hexadecimal()
		var userIn jdbcapp.AppResponseUser
		app.POSTf("", "?name=%s", value).ParseInto(&userIn)

		By("updating the service to set 'use_managed_admin_password' a first time which is expected to fail")
		params := `{"use_managed_admin_password": true}`
		session := cf.Start("update-service", serviceInstance.Name, "-c", params, "--wait")
		Eventually(session).WithTimeout(time.Hour).Should(gexec.Exit(1), func() string {
			out, _ := cf.Run("service", serviceInstance.Name)
			return out
		})

		By("checking that it fails for the expected reason")
		msg, _ := cf.Run("service", serviceInstance.Name)
		Expect(msg).To(MatchRegexp(`message:\s+update failed:\s+Error:\s+Provider produced inconsistent final plan When expanding the plan for aws_secretsmanager_secret_rotation`))

		By("updating the service to set 'use_managed_admin_password' a second time")
		serviceInstance.Update(services.WithParameters(params))

		By("waiting for the password rotation to be applied")
		identifier := fmt.Sprintf("csb-mssql-%s", serviceInstance.GUID())
		Eventually(func() string {
			status := dbInstanceStatus(identifier)
			Expect(status).To(SatisfyAny(Equal("resetting-master-credentials"), Equal("available")))
			return status
		}).WithTimeout(time.Hour).WithPolling(time.Minute).Should(Equal("available"))

		By("rebinding app")
		binding.Unbind()
		serviceInstance.Bind(app)
		app.Restage()

		By("getting the previously stored value")
		var userOut jdbcapp.AppResponseUser
		app.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut.Name)

		By("updating the service to unset 'use_managed_admin_password'")
		serviceInstance.Update(services.WithParameters(`{"use_managed_admin_password": false}`))

		By("rebinding app")
		binding.Unbind()
		serviceInstance.Bind(app)
		app.Restage()

		By("getting the previously stored value")
		app.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut.Name)
	})
})
