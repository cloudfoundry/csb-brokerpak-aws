package upgrade_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("UpgradePostgreSQLTest", Label("postgresql", "upgrade"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			const (
				postgreSQLPlanToUpgradeEngine = `[{"name":"default_postgres_version14","id":"77de3441-1096-48aa-8909-a7dc5e457fa2","description":"Default Postgres plan with version 14.x","display_name":"default_postgres_version14.x","instance_class":"db.m6i.large","postgres_version":"14","storage_gb":100},{"name":"default_postgres_version13","id":"95989511-5e6f-4845-ae26-1401e077c193","description":"Default Postgres plan with version 13.10","display_name":"default_postgres_version13","instance_class":"db.m6i.large","postgres_version":"13.10","storage_gb":100}]`
				plansVar                      = `GSB_SERVICE_CSB_AWS_POSTGRESQL_PLANS`
			)

			customPlans := apps.EnvVar{Name: plansVar, Value: postgreSQLPlanToUpgradeEngine}

			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-postgresql"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
				brokers.WithEnv(customPlans),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-aws-postgresql",
				services.WithPlan("default_postgres_version13"),
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.PostgreSQL), apps.WithDisk("2G"))
			appTwo := apps.Push(apps.WithApp(apps.PostgreSQL), apps.WithDisk("2G"))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

			By("creating a schema using the first app")
			schema := random.Name(random.WithMaxLength(10))
			appOne.PUT("", schema)

			By("setting a key-value using the first app")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
			appOne.PUT(valueOne, "%s/%s", schema, keyOne)

			By("getting the value using the second app")
			got := appTwo.GET("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir, customPlans)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("checking previously written data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("updating the instance plan")
			serviceInstance.Update(services.WithPlan("default_postgres_version14"))

			By("checking previously written data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking previously written data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne).String()
			Expect(got).To(Equal(valueOne))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, "%s/%s", schema, keyTwo)

			got = appTwo.GET("%s/%s", schema, keyTwo).String()
			Expect(got).To(Equal(valueTwo))
		})
	})
})
