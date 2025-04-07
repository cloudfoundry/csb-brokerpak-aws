package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/awscli"
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"csbbrokerpakaws/acceptance-tests/helpers/jdbcapp"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
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
		appOne.POSTf("", "?name=%s", value).ParseInto(&userIn)

		By("binding and starting the second app")
		serviceInstance.Bind(appTwo)
		apps.Start(appTwo)

		By("getting the entry using the second app")
		appTwo.GETf("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

		By("verifying the DB connection utilises TLS")
		appOne.GET("mysql-ssl").ParseInto(&sslInfo)
		Expect(strings.ToLower(sslInfo.VariableName)).To(Equal("ssl_cipher"))
		Expect(sslInfo.Value).NotTo(BeEmpty())

		By("deleting the entry using the first app")
		appOne.DELETEf("%d", userIn.ID)

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
		response := golangApp.GETResponsef("%s?tls=false", key)
		defer response.Body.Close()
		Expect(response).To(HaveHTTPStatus(http.StatusInternalServerError), "force TLS is enabled by default")
		b, err := io.ReadAll(response.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body in TLS failure")
		Expect(string(b)).To(ContainSubstring("error connecting to database: failed to verify the connection"), "force TLS is enabled by default")
		Expect(string(b)).To(ContainSubstring("Error 1045 (28000): Access denied for user"), "postgresql client cannot connect to the postgres server due to invalid TLS")
	})

	// As we introduce the 'use_managed_admin_password' feature, some users may wish to update existing DBs.
	// This is a tactical test that should exist for this changeover period and is not intended to be a forever test.
	// Due to limitations in Tofu/AWS provider/AWS the operation to switch fails first time, then succeeds on
	// a second attempt. That's not an ideal customer experience, and this test exists to ensure that what we
	// document works, and make us aware if the behavior changes.
	It("allows 'use_managed_admin_password' to be enabled", Label("managed-password"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-mysql", services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing an unstarted app")
		appManifest := jdbcapp.ManifestFor(jdbcapp.MySQL)
		app := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(appManifest))
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
		const params = `{"use_managed_admin_password": true}`
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
		identifier := fmt.Sprintf("csb-mysql-%s", serviceInstance.GUID())
		Eventually(dbInstanceStatus(identifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

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

	// While snapshot restore is an AWS feature rather than a CSB feature, its valuable to check that it works
	It("allows a snapshot to be restored when 'use_managed_admin_password' is enabled", Label("managed-password-snapshot-restore"), func() {
		By("creating a service instance")
		const params = `{"use_managed_admin_password":true, "rotate_admin_password_after":4}`
		serviceInstance := services.CreateInstance("csb-aws-mysql", services.WithPlan("default"), services.WithParameters(params))
		defer serviceInstance.Delete()

		By("pushing an unstarted app")
		appManifest := jdbcapp.ManifestFor(jdbcapp.MySQL)
		app := apps.Push(apps.WithApp(apps.JDBCTestAppMysql), apps.WithManifest(appManifest))
		defer apps.Delete(app)

		By("waiting for the DB to be available")
		dbInstanceIdentifier := fmt.Sprintf("csb-mysql-%s", serviceInstance.GUID())
		Eventually(dbInstanceStatus(dbInstanceIdentifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("binding the app to the service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating an entry using the app")
		value := random.Hexadecimal()
		var userIn jdbcapp.AppResponseUser
		app.POSTf("", "?name=%s", value).ParseInto(&userIn)

		By("taking a snapshot of the DB")
		snapshotIdentifier := random.Name(random.WithPrefix("snapshot-restore-test"))
		awscli.AWS(
			"rds", "create-db-snapshot",
			"--db-snapshot-identifier", snapshotIdentifier,
			"--db-instance-identifier", dbInstanceIdentifier,
		)
		Eventually(dbSnapshotStatus(snapshotIdentifier)).WithPolling(time.Minute).WithTimeout(time.Hour).Should(Equal("available"))

		By("deleting the DB")
		awscli.AWS("rds", "delete-db-instance", "--db-instance-identifier", dbInstanceIdentifier, "--skip-final-snapshot")
		Eventually(func() *gbytes.Buffer {
			session := awscli.AWSSession("rds", "describe-db-instances", "--db-instance-identifier", dbInstanceIdentifier, "--query", "DBInstances[0].DBInstanceStatus")
			session.Wait(5 * time.Minute) // Should be quick, but it's occasionally slow, and we don't want to bail out when that happens
			return session.Err
		}).WithPolling(time.Minute).WithTimeout(time.Hour).Should(gbytes.Say("not found"))

		By("waiting a bit to avoid a 'DBInstanceAlreadyExists' error if we restore too soon")
		time.Sleep(time.Minute)

		// At the time of writing, there's no flag in the AWS CLI (or in the console) to enable AWS secrets manager when
		// restoring from a snapshot. Ideally we could restore with the "--manage-master-user-password" and there would
		// be no need to mess around with the settings after the restore, but that flag hasn't been added yet.
		By("restoring the DB from the snapshot")
		awscli.AWS(
			"rds", "restore-db-instance-from-db-snapshot",
			"--db-snapshot-identifier", snapshotIdentifier,
			"--db-instance-identifier", dbInstanceIdentifier,
			"--db-instance-class", "db.t3.micro",
		)

		By("waiting for the DB to be available")
		Eventually(dbInstanceStatus(dbInstanceIdentifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("disabling 'use_managed_admin_password' to match the restored snapshot")
		serviceInstance.Update(services.WithParameters(`{"use_managed_admin_password": false}`))

		By("updating the service to set 'use_managed_admin_password' a first time which is expected to fail")
		session := cf.Start("update-service", serviceInstance.Name, "-c", params, "--wait")
		Eventually(session).WithTimeout(time.Hour).Should(gexec.Exit(1), func() string {
			out, _ := cf.Run("service", serviceInstance.Name)
			return out
		})

		By("checking that it failed for the expected reason")
		msg, _ := cf.Run("service", serviceInstance.Name)
		Expect(msg).To(MatchRegexp(`message:\s+update failed:\s+Error:\s+Provider produced inconsistent final plan When expanding the plan for aws_secretsmanager_secret_rotation`))

		By("updating the service to set 'use_managed_admin_password' a second time")
		serviceInstance.Update(services.WithParameters(params))

		By("waiting for the password rotation to be applied")
		Eventually(dbInstanceStatus(dbInstanceIdentifier)).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("checking that bindings created before the restore still work")
		var userOut1 jdbcapp.AppResponseUser
		app.GETf("%d", userIn.ID).ParseInto(&userOut1)
		Expect(userOut1.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut1.Name)

		By("rebinding app")
		binding.Unbind()
		serviceInstance.Bind(app)
		app.Restage()

		By("checking that new bindings can be created")
		var userOut2 jdbcapp.AppResponseUser
		app.GETf("%d", userIn.ID).ParseInto(&userOut2)
		Expect(userOut2.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut2.Name)
	})
})

func dbSnapshotStatus(id string) func() string {
	return func() string {
		status := awscli.AWSQuery("DBSnapshots[0].Status", "rds", "describe-db-snapshots", "--db-snapshot-identifier", id)
		Expect(status).To(SatisfyAny(Equal("available"), Equal("creating")))
		return status
	}
}
