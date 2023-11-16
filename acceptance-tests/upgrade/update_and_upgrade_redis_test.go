package upgrade_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("Redis", Label("redis"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-aws-redis"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			serviceInstance := services.CreateInstance(
				"csb-aws-redis",
				services.WithPlan("example-with-flexible-node-type"),
				services.WithParameters(map[string]any{"node_type": "cache.t3.medium"}),
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.Redis))
			appTwo := apps.Push(apps.WithApp(apps.Redis))
			defer apps.Delete(appOne, appTwo)

			By("binding the apps to the Redis service instance")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

			By("setting a key-value using the first app")
			key := random.Hexadecimal()
			value := random.Hexadecimal()
			appOne.PUT(value, "/primary/%s", key)

			By("getting the value using the second app")
			Expect(appTwo.GET("/primary/%s", key).String()).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the value using the second app")
			Expect(appTwo.GET("/primary/%s", key).String()).To(Equal(value))

			By("updating the instance plan")
			serviceInstance.Update(services.WithParameters(map[string]any{"node_type": "cache.t3.medium"}))

			By("getting the value using the second app")
			Expect(appTwo.GET("/primary/%s", key).String()).To(Equal(value))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("getting the value using the second app")
			Expect(appTwo.GET("/primary/%s", key).String()).To(Equal(value))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, "/primary/%s", keyTwo)
			Expect(appTwo.GET("/primary/%s", keyTwo).String()).To(Equal(valueTwo))
		})
	})

})
