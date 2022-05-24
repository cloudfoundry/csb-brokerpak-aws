package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DynamoDB", Label("dynamodb"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-dynamodb",
			"ondemand",
			services.WithParameters(config()),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.DynamoDB))
		appTwo := apps.Push(apps.WithApp(apps.DynamoDB))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the DynamoDB service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(HaveKey("credhub-ref"))

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
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
		"table_name": random.Name(random.WithPrefix("csb", "dynamodb")),
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
