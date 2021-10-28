package dynamodb_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("DynamoDB", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := helpers.CreateService("csb-aws-dynamodb", "ondemand", config())
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.DynamoDB)
		appTwo := helpers.AppPushUnstarted(apps.DynamoDB)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the DynamoDB service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(HaveKey("credhub-ref"))

		By("setting a key-value using the first app")
		key := helpers.RandomHex()
		value := helpers.RandomHex()
		appOne.PUT(value, key)

		By("getting the value using the second app")
		got := appTwo.GET(key)
		Expect(got).To(Equal(value))
	})
})

func config() interface{} {
	return map[string]interface{}{
		"attributes": []map[string]interface{}{
			{
				"name": "id",
				"type": "S",
			},
			{
				"name": "key",
				"type": "S",
			},
			{
				"name": "value",
				"type": "S",
			},
		},
		"hash_key":   "id",
		"range_key":  "value",
		"table_name": helpers.RandomName("csb", "dynamodb"),
		"global_secondary_indexes": []map[string]interface{}{
			{
				"name":               "KeyIndex",
				"hash_key":           "key",
				"range_key":          "value",
				"projection_type":    "INCLUDE",
				"non_key_attributes": []string{"id"},
			},
		},
	}
}
