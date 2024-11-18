package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/awscli"
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/jdbcapp"
	"csbbrokerpakaws/acceptance-tests/helpers/matchers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("PostgreSQL", Label("postgresql"), func() {
	It("can be accessed by an app", Label("JDBC-p"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-postgresql", services.WithPlan("default"))
		defer serviceInstance.Delete()

		postgresTestMultipleApps(serviceInstance)
	})

	It("works with latest changes to public schema in postgres 15", Label("Postgres15"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-postgresql", services.WithPlan("pg15"))
		defer serviceInstance.Delete()

		postgresTestMultipleApps(serviceInstance)
	})

	It("works with postgres 16", Label("Postgres16"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-postgresql", services.WithPlan("pg16"))
		defer serviceInstance.Delete()

		postgresTestMultipleApps(serviceInstance)
	})

	// As we introduce the 'use_managed_admin_password' feature, some users may wish to update existing DBs.
	// This is a tactical test that should exist for this changeover period and is not intended to be a forever test.
	// Due to limitations in Tofu/AWS provider/AWS the operation to switch fails first time, then succeeds on
	// a second attempt. That's not an ideal customer experience, and this test exists to ensure that what we
	// document works, and make us aware if the behavior changes.
	It("allows 'use_managed_admin_password' to be enabled", Label("managed-password"), func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-postgresql", services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing an unstarted app")
		appManifest := jdbcapp.ManifestFor(jdbcapp.PostgreSQL)
		app := apps.Push(apps.WithApp(apps.JDBCTestAppPostgres), apps.WithManifest(appManifest))
		defer apps.Delete(app)

		By("binding the app to the service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating an entry using the app")
		value := random.Hexadecimal()
		var userIn jdbcapp.AppResponseUser
		app.POST("", "?name=%s", value).ParseInto(&userIn)

		By("updating the service to set 'use_managed_admin_password' a first time which is expected to fail")
		params := `{"use_managed_admin_password": true}`
		session := cf.Start("update-service", serviceInstance.Name, "-c", params, "--wait")
		Eventually(session).WithTimeout(time.Hour).Should(gexec.Exit(), func() string {
			out, _ := cf.Run("service", serviceInstance.Name)
			return out
		})

		By("checking that it fails for the expected reason")
		msg, _ := cf.Run("service", serviceInstance.Name)
		Expect(msg).To(MatchRegexp(`message:\s+update failed:\s+Error:\s+Provider produced inconsistent final plan When expanding the plan for aws_secretsmanager_secret_rotation`))

		By("updating the service to set 'use_managed_admin_password' a second time")
		serviceInstance.Update(services.WithParameters(params))

		By("waiting for the password rotation to be applied")
		identifier := fmt.Sprintf("csb-postgresql-%s", serviceInstance.GUID())
		Eventually(func() string {
			status := dbInstanceStatus(identifier)
			Expect(status).To(SatisfyAny(Equal("resetting-master-credentials"), Equal("available")))
			return status
		}).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Equal("available"))

		By("rebinding app")
		binding.Unbind()
		serviceInstance.Bind(app)

		By("getting the previously stored value")
		var userOut jdbcapp.AppResponseUser
		app.GET("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut.Name)

		By("updating the service to unset 'use_managed_admin_password'")
		serviceInstance.Update(services.WithParameters(`{"use_managed_admin_password": false}`))

		By("rebinding app")
		binding.Unbind()
		serviceInstance.Bind(app)

		By("getting the previously stored value")
		app.GET("%d", userIn.ID).ParseInto(&userOut)
		Expect(userOut.Name).To(Equal(value), "App stored [%s] as the value, App retrieved [%s]", value, userOut.Name)
	})
})

func postgresTestMultipleApps(serviceInstance *services.ServiceInstance) {
	var (
		userIn, userOut jdbcapp.AppResponseUser
		sslInfo         jdbcapp.PostgresSSLInfo
	)

	By("pushing the unstarted app")
	appManifest := jdbcapp.ManifestFor(jdbcapp.PostgreSQL)
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
	appOne.POST("", "?name=%s", value).ParseInto(&userIn)

	By("binding and starting the second app")
	serviceInstance.Bind(appTwo)
	apps.Start(appTwo)

	By("getting the entry using the second app")
	appTwo.GET("%d", userIn.ID).ParseInto(&userOut)
	Expect(userOut.Name).To(Equal(value), "The first app stored [%s] as the value, the second app retrieved [%s]", value, userOut.Name)

	By("verifying the DB connection utilises TLS")
	appOne.GET("postgres-ssl").ParseInto(&sslInfo)
	Expect(sslInfo.SSL).To(BeTrue())
	Expect(sslInfo.Cipher).NotTo(BeEmpty())
	Expect(sslInfo.Bits).To(BeNumerically(">=", 256))

	By("deleting the entry using the first app")
	appOne.DELETE("%d", userIn.ID)
}

func dbInstanceStatus(instanceName string) string {
	var receiver struct {
		Status []string `jsonry:"DBInstances.DBInstanceStatus"`
	}
	awscli.AWSToJSON(&receiver, "rds", "describe-db-instances", "--db-instance-identifier", instanceName)
	Expect(receiver.Status).To(HaveLen(1))
	return receiver.Status[0]
}
