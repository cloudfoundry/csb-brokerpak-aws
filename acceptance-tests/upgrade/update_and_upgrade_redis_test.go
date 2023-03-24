package upgrade_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
				services.WithPlan("small"),
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
			Expect(appTwo.GET("/primary/%s", key)).To(Equal(value))

			By("pushing the development version of the broker")
			const (
				plansRedisVar              = "GSB_SERVICE_CSB_AWS_REDIS_PLANS"
				defaultRedisPlanForUpgrade = `{"name":"default","id":"c7f64994-a1d9-4e1f-9491-9d8e56bbf146","description":"Default Redis plan","display_name":"default","node_type":"cache.t3.medium","redis_version":"7.0", "at_rest_encryption_enabled": false, "multi_az_enabled": false, "automatic_failover_enabled": false}`
				oldRedisPlanForUpgrade     = `{"name": "small", "id": "ad963fcd-19f7-4b79-8e6d-645756e84f7a","description": "Beta - Redis 6.0 with 1GB cache and 1 node.","cache_size": 2,"redis_version": "6.0","node_count": 1,"at_rest_encryption_enabled": false, "multi_az_enabled": false, "automatic_failover_enabled": false}`
			)
			planOverride := apps.EnvVar{Name: plansRedisVar, Value: fmt.Sprintf(`[%s, %s]`, defaultRedisPlanForUpgrade, oldRedisPlanForUpgrade)}
			serviceBroker.UpdateBroker(developmentBuildDir, planOverride)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the value using the second app")
			Expect(appTwo.GET("/primary/%s", key)).To(Equal(value))

			By("updating the instance plan")
			serviceInstance.Update(services.WithPlan("default"))

			By("getting the value using the second app")
			Expect(appTwo.GET("/primary/%s", key)).To(Equal(value))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("getting the value using the second app")
			Expect(appTwo.GET("/primary/%s", key)).To(Equal(value))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, "/primary/%s", keyTwo)
			Expect(appTwo.GET("/primary/%s", keyTwo)).To(Equal(valueTwo))
		})
	})

})
