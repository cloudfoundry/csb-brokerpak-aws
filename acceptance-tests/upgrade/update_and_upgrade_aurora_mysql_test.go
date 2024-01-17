package upgrade_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("UpgradeAuroraMySQLTest", Label("aurora-mysql", "upgrade"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-aurora-mysql"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			// The auto minor version upgrade is enabled by default.
			// Performing tests using the major version for aurora-mysql 8, in other words, version 8.0
			// we received this error:
			// Error: creating RDS Cluster (csb-auroramysql-xxxx):
			// InvalidParameterCombination: Cannot find version 8.0.mysql_aurora.3.04.0 for aurora-mysql status code: 400
			//
			// The `8.0.mysql_aurora.3.04.0` version does not appear in the AWS console or by running the following commands:
			// `aws rds describe-db-engine-versions --engine aurora-mysql --engine-version 8.0 --region us-west-2`
			// or `aws rds describe-db-engine-versions --engine aurora-mysql --output text --region us-west-2`
			// There is no open incidence in the provider mentioning anything about it.
			// We change the test and proceed to document issue.
			serviceInstance := services.CreateInstance(
				"csb-aws-aurora-mysql",
				services.WithPlan("default"),
				services.WithParameters(map[string]any{
					"cluster_instances": 1,
					"instance_class":    "db.t3.medium",
					"engine_version":    "5.7",
				}),
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
			Expect(appTwo.GET(key).String()).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the value using the second app")
			Expect(appTwo.GET(key).String()).To(Equal(value))

			By("updating the instance")
			serviceInstance.Update(services.WithParameters(map[string]any{"cluster_instances": 2}))

			By("getting the value using the second app")
			Expect(appTwo.GET(key).String()).To(Equal(value))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("getting the value using the second app")
			Expect(appTwo.GET(key).String()).To(Equal(value))

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, keyTwo)
			Expect(appTwo.GET(keyTwo).String()).To(Equal(valueTwo))
		})
	})
})
