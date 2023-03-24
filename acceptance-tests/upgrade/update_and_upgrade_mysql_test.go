package upgrade_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("UpgradeMySQLTest", Label("mysql", "upgrade"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			const (
				mysql57Plans = `[{"name":"default-5.7","id":"ce70e430-5a08-11ed-a801-367dda7ea869","description":"DefaultMySQL 5.7 plan","display_name":"default-5.7","instance_class":"db.t3.medium","mysql_version":"5.7","storage_gb":10,"storage_encrypted":false,"multi_az":false,"storage_autoscale":false,"storage_type":"gp2"},{"name":"small","id":"2268ce43-7fd7-48dc-be2f-8611e11fb12e","description":"MySQLv5.7, minimum 2 cores, minimum 4GB ram, 5GB storage","display_name":"small","storage_gb":5,"storage_type":"gp2","cores":2,"mysql_version":"5.7","storage_encrypted":false,"multi_az":false,"storage_autoscale":false}]`
				plansVar     = `GSB_SERVICE_CSB_AWS_MYSQL_PLANS`
			)

			customPlans := apps.EnvVar{Name: plansVar, Value: mysql57Plans}

			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-aws-mysql"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
				brokers.WithEnv(customPlans),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-aws-mysql",
				services.WithPlan("small"),
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MySQL))
			appTwo := apps.Push(apps.WithApp(apps.MySQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

			By("setting a key-value using the first app")
			key := random.Hexadecimal()
			value := random.Hexadecimal()
			appOne.PUT(value, key)
			By("getting the value using the second app")
			Expect(appTwo.GET(key)).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir, customPlans)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the value using the second app")
			Expect(appTwo.GET(key)).To(Equal(value))

			By("updating the instance plan")
			serviceInstance.Update(services.WithPlan("default-5.7"))

			By("getting the value using the second app")
			Expect(appTwo.GET(key)).To(Equal(value))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("getting the value using the second app")
			Expect(appTwo.GET(key)).To(Equal(value))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, keyTwo)
			Expect(appTwo.GET(keyTwo)).To(Equal(valueTwo))
		})
	})
})
