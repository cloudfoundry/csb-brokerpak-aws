package upgrade_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/plans"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("UpgradeMSSQLTest", Label("mssql", "upgrade"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")

			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-aws-mssql"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceOffering := "csb-aws-mssql"
			servicePlan := "default"
			serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
			// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
			// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
			// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
			// (which can only have 5 instances at any time) to prevent subsequent test failures.
			defer services.Delete(serviceName)
			// without backups and multi az to speed up the process
			params := map[string]any{
				"backup_retention_period": 0,
				"multi_az":                false,
			}

			serviceInstance := services.CreateInstance(
				serviceOffering,
				services.WithPlan(servicePlan),
				services.WithParameters(params),
				services.WithBroker(serviceBroker),
				services.WithName(serviceName),
			)

			By("pushing the unstarted app twice")
			appWriter := apps.Push(apps.WithApp(apps.MSSQL))
			appReader := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appWriter, appReader)

			By("binding the the writer app")
			bindingWriter := serviceInstance.Bind(appWriter)

			By("starting the writer app")
			apps.Start(appWriter)

			By("creating a schema using the first app")
			schema := random.Name(random.WithMaxLength(10))
			appWriter.PUTf("", "%s?dbo=false", schema)

			By("setting a key-value using the first app")
			key := random.Hexadecimal()
			value := random.Hexadecimal()
			appWriter.PUTf(value, "%s/%s", schema, key)

			By("binding the reader app")
			bindingReader := serviceInstance.Bind(appReader)

			By("starting the reader app")
			apps.Start(appReader)

			By("getting the entry using the reader app")
			got := appReader.GETf("%s/%s", schema, key).String()
			Expect(got).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable(servicePlan, serviceOffering, serviceBroker.Name))

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the entry using the reader app after upgrading")
			got = appReader.GETf("%s/%s", schema, key).String()
			Expect(got).To(Equal(value))

			By("deleting bindings created before the upgrade")
			bindingWriter.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appWriter)
			apps.Restage(appWriter)

			By("updating service instance")
			serviceInstance.Update(services.WithParameters(`{}`))

			By("deleting bindings created before the update")
			bindingWriter.Unbind()
			bindingReader.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appWriter)
			serviceInstance.Bind(appReader)
			apps.Restage(appWriter, appReader)

			By("checking data can still be written and read")
			key = random.Hexadecimal()
			value = random.Hexadecimal()
			appWriter.PUTf(value, "%s/%s", schema, key)
			got = appReader.GETf("%s/%s", schema, key).String()
			Expect(got).To(Equal(value))
		})
	})
})
