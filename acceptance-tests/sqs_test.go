package acceptance_tests_test

import (
	"net/http"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SQS", Label("sqs"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-sqs", services.WithPlan("standard"))
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.SQS))
		appTwo := apps.Push(apps.WithApp(apps.SQS))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		bindingOneName := random.Name(random.WithPrefix("producer"))
		binding := serviceInstance.Bind(appOne, services.WithBindingName(bindingOneName))
		bindingTwoName := random.Name(random.WithPrefix("consumer"))
		serviceInstance.Bind(appTwo, services.WithBindingName(bindingTwoName))

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(HaveKey("credhub-ref"))

		By("sending a message using the first app")
		message := random.Hexadecimal()
		appOne.POST(message, "/send/%s", bindingOneName)

		By("receiving the message using the second app")
		got := appTwo.GET("/retrieve_and_delete/%s", bindingTwoName).String()
		Expect(got).To(Equal(message))
	})

	It("should work with FIFO and DLQ", func() {
		By("creating a FIFO DLQ service instance")
		dlqServiceInstance := services.CreateInstance(
			"csb-aws-sqs",
			services.WithPlan("fifo"),
			services.WithParameters(map[string]any{"dlq": true}),
		)
		defer dlqServiceInstance.Delete()

		csbKey := dlqServiceInstance.CreateServiceKey()
		defer csbKey.Delete()
		var skReceiver struct {
			ARN string `json:"arn"`
		}
		csbKey.Get(&skReceiver)

		By("creating a FIFO Queue")
		standardQueueServiceInstance := services.CreateInstance(
			"csb-aws-sqs",
			services.WithPlan("fifo"),
			services.WithParameters(map[string]any{
				"dlq_arn":           skReceiver.ARN,
				"max_receive_count": 1,
			}),
		)
		defer standardQueueServiceInstance.Delete()

		By("pushing the unstarted apps")
		producerApp := apps.Push(apps.WithName(random.Name(random.WithPrefix("producer"))), apps.WithApp(apps.SQS))
		consumerApp := apps.Push(apps.WithName(random.Name(random.WithPrefix("consumer"))), apps.WithApp(apps.SQS))
		defer apps.Delete(producerApp, consumerApp)

		By("binding the apps to the service instance - producer/consumer")
		producerBindingName := random.Name(random.WithPrefix("producer"))
		producerBinding := standardQueueServiceInstance.Bind(producerApp, services.WithBindingName(producerBindingName))
		defer producerBinding.Unbind()

		consumerBindingName := random.Name(random.WithPrefix("consumer"))
		consumerBinding := standardQueueServiceInstance.Bind(consumerApp, services.WithBindingName(consumerBindingName))
		defer consumerBinding.Unbind()

		By("binding the app to the service instance - DLQ consumer")
		bindingDLQName := random.Name(random.WithPrefix("dlq"))
		dlqBinding := dlqServiceInstance.Bind(consumerApp, services.WithBindingName(bindingDLQName))
		defer dlqBinding.Unbind()

		By("starting the apps")
		apps.Start(producerApp, consumerApp)

		By("sending a message - producer")
		message := random.Hexadecimal()
		producerApp.POST(message, "/send/%s", producerBindingName)

		By("read a message without delete it - consumer")
		got := consumerApp.GET("/retrieve/%s", consumerBindingName).String()
		Expect(got).To(Equal(message))

		By("attempts retrieving from the queue again so transferring the message to the DLQ is triggerd - consumer")
		response := consumerApp.GETResponse("/retrieve/%s", consumerBindingName)
		Expect(response).To(HaveHTTPStatus(http.StatusTooEarly))

		By("reading message from DLQ - DLQ consumer")
		got = consumerApp.GET("/retrieve_and_delete/%s", bindingDLQName).String()
		Expect(got).To(Equal(message))
	})
})
