package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DynamoDB Table", Label("dynamodb-table"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-dynamodb-table",
			services.WithPlan("ondemand"),
			services.WithParameters(config()),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.DynamoDBTable))
		appTwo := apps.Push(apps.WithApp(apps.DynamoDBTable))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the DynamoDB Table service instance")
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
		got := appTwo.GET(key).String()
		Expect(got).To(Equal(value))
	})
})

func config() any {
	return map[string]any{
		"attributes": []map[string]any{
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
		"global_secondary_indexes": []map[string]any{
			{
				"name":               "KeyIndex",
				"hash_key":           "key",
				"range_key":          "value",
				"projection_type":    "INCLUDE",
				"non_key_attributes": []string{"id"},
			},
		},
		"iam_arn":        "arn:aws:iam::000000000000:user/this-user-must-exist-and-have-permission-to-assume-role",
		"role_name":      "this-role-must-exist-and-allow-the-above-user-to-assume-role",
		"region":         "us-west-1",
		"creds_endpoint": "https://the-endpoint-for-the-cloud-service-broker.csb.cf-app.com/creds",
	}
}
