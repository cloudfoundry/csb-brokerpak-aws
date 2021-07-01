package redis_test

import (
	"acceptancetests/helpers"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Redis", func() {
	It("can be accessed by an app", func() {
		By("building the app")
		appDir := helpers.BuildApp("./redisapp")
		defer os.RemoveAll(appDir)

		By("pushing the unstarted app twice")
		names := helpers.PushAppsUnstarted("redis", appDir, 2)
		defer helpers.DeleteApps(names)

		By("binding the apps to the Redis service instance")
		bindingName := helpers.Bind(names[0], serviceInstanceName)
		helpers.Bind(names[1], serviceInstanceName)

		By("starting the apps")
		helpers.StartApps(names)

		By("checking that the app environment has a credhub reference for credentials")
		creds := helpers.GetBindingCredential(names[0], "csb-aws-redis", bindingName)
		Expect(creds).To(HaveKey("credhub-ref"))

		By("setting a key using the first app")
		key := helpers.RandomString()
		value := helpers.RandomString()
		helpers.HTTPPost(fmt.Sprintf("http://%s.%s/%s", names[0], helpers.DefaultSharedDomain(), key), value)

		By("getting the key using the second app")
		got := helpers.HTTPGet(fmt.Sprintf("http://%s.%s/%s", names[1], helpers.DefaultSharedDomain(), key))
		Expect(got).To(Equal(value))
	})
})
