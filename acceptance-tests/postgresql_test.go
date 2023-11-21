package acceptance_tests_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
