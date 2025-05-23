package upgrade_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/plans"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeSQSTest", Label("upgrade", "sqs"), func() {
	Context("When upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-sqs"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceOffering := "csb-aws-sqs"
			servicePlan := "fifo"
			serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
			// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
			// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
			// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
			// (which can only have 5 instances at any time) to prevent subsequent test failures.
			defer services.Delete(serviceName)
			serviceInstance := services.CreateInstance(
				serviceOffering,
				services.WithPlan(servicePlan),
				services.WithBroker(serviceBroker),
				services.WithName(serviceName),
			)

			By("pushing the unstarted app")
			app := apps.Push(apps.WithApp(apps.SQS))
			defer apps.Delete(app)

			By("binding the app to the service instance")
			bindingName := random.Name(random.WithPrefix("binding"))
			binding := serviceInstance.Bind(app, services.WithBindingName(bindingName))
			apps.Start(app)

			By("sending two messages")
			send := func(message string) {
				messageGroupID := random.Hexadecimal()
				messageDeduplicationID := random.Hexadecimal()
				app.POSTf(message, "/send/%s?messageGroupId=%s&messageDeduplicationId=%s", bindingName, messageGroupID, messageDeduplicationID)
			}
			messageOne := random.Hexadecimal()
			messageTwo := random.Hexadecimal()
			send(messageOne)
			send(messageTwo)

			By("pushing the development version of the broker")
			serviceBroker.UpdateBroker(developmentBuildDir)

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable(servicePlan, serviceOffering, serviceBroker.Name))

			By("upgrading the service instance")
			serviceInstance.Upgrade()

			By("receiving the previously written first message")
			got := app.GETf("/retrieve_and_delete/%s", bindingName).String()
			Expect(got).To(Equal(messageOne))

			By("deleting bindings created before the upgrade")
			binding.Unbind()

			By("binding the app to the instance again")
			serviceInstance.Bind(app, services.WithBindingName(bindingName))
			apps.Restage(app)

			By("updating the service instance")
			serviceInstance.Update(services.WithParameters(`{}`))

			By("receiving the previously written second message")
			got = app.GETf("/retrieve_and_delete/%s", bindingName).String()
			Expect(got).To(Equal(messageTwo))

			By("deleting bindings created before the upgrade")
			binding.Unbind()

			By("binding the app to the instance again")
			serviceInstance.Bind(app, services.WithBindingName(bindingName))
			apps.Restage(app)

			By("checking that messages can still be produced and consumed")
			messageThree := random.Hexadecimal()
			send(messageThree)

			got = app.GETf("/retrieve_and_delete/%s", bindingName).String()
			Expect(got).To(Equal(messageThree))
		})
	})
})
