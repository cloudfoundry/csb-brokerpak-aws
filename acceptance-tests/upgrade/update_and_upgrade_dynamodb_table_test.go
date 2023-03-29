package upgrade_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("UpgradeDynamoDBTableTest", Label("dynamodb-table", "upgrade"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")

			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-aws-dynamodb-table"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			tableName := random.Name(random.WithPrefix("csb", "dynamodb"))
			serviceInstance := services.CreateInstance(
				"csb-aws-dynamodb", // old service offering name
				services.WithPlan("ondemand"),
				services.WithParameters(config(tableName)),
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.DynamoDBTable))
			appTwo := apps.Push(apps.WithApp(apps.DynamoDBTable))
			defer apps.Delete(appOne, appTwo)

			By("binding the apps to the DynamoDB Table service instance")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

			By("setting a key-value using the first app")
			key := random.Hexadecimal()
			value := random.Hexadecimal()
			appOne.PUT(value, key)

			By("getting the value using the second app")
			got := appTwo.GET(key)
			Expect(got).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the value using the second app")
			Expect(appTwo.GET(key)).To(Equal(value))

			By("updating the instance plan")
			serviceInstance.Update(services.WithParameters(updatedConfig(tableName)))

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

func config(tableName string) any {
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
		"table_name": tableName,
		"global_secondary_indexes": []map[string]any{
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

func updatedConfig(tableName string) any {
	return map[string]any{
		"server_side_encryption_enabled": true,
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
		"table_name": tableName,
		"global_secondary_indexes": []map[string]any{
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
