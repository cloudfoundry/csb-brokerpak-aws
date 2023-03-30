package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Redis", Label("redis"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-redis", services.WithPlan("default"))
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.Redis))
		appTwo := apps.Push(apps.WithApp(apps.Redis))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the Redis service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(HaveKey("credhub-ref"))

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appOne.PUT(value, "/primary/%s", key)

		By("getting the value using the second app")
		got := appTwo.GET("/primary/%s", key).String()
		Expect(got).To(Equal(value))

		By("getting the value using the reader endpoint")
		got2 := appTwo.GET("/replica/%s", key).String()
		Expect(got2).To(Equal(value))
	})
})
