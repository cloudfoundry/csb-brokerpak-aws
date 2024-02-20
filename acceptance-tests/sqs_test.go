package acceptance_tests_test

import (
	"io"
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
		got := appTwo.GET("/receive/%s", bindingTwoName).String()
		Expect(got).To(Equal(message))
	})

	FIt("should work with DLQ", func() {
		By("creating a DLS service instance")
		// dlqServiceInstance := services.CreateInstance(
		// 	"csb-aws-sqs",
		// 	services.WithPlan("standard"),
		// 	services.WithParameters(map[string]any{"dlq": true}),
		// )
		dlqServiceInstance := services.ServiceInstance{Name: "dlq1"}
		// defer dlqServiceInstance.Delete()

		// csbKey := dlqServiceInstance.CreateServiceKey()
		// defer csbKey.Delete()
		// var skReceiver struct {
		// 	ARN string `json:"arn"`
		// }
		// csbKey.Get(&skReceiver)

		By("creating a Standard Queue")
		// standardQueueServiceInstance := services.CreateInstance(
		// 	"csb-aws-sqs",
		// 	services.WithPlan("standard"),
		// 	services.WithParameters(map[string]any{
		// 		"dlq_arn":           skReceiver.ARN,
		// 		"max_receive_count": 5,
		// 	}),
		// )
		standardQueueServiceInstance := services.ServiceInstance{Name: "producer1"}
		// defer standardQueueServiceInstance.Delete()

		By("pushing the unstarted apps")
		producerApp := apps.Push(apps.WithName(random.Name(random.WithPrefix("producer"))), apps.WithApp(apps.SQS))
		consumerApp := apps.Push(apps.WithName(random.Name(random.WithPrefix("consumer"))), apps.WithApp(apps.SQS))
		dlqConsumerApp := apps.Push(apps.WithName(random.Name(random.WithPrefix("dlq"))), apps.WithApp(apps.SQS))
		defer apps.Delete(producerApp, consumerApp, dlqConsumerApp)

		By("binding the apps to the service instance - producer/consumer")
		producerBindingName := random.Name(random.WithPrefix("producer"))
		producerBinding := standardQueueServiceInstance.Bind(producerApp, services.WithBindingName(producerBindingName))
		defer producerBinding.Unbind()

		consumerBindingName := random.Name(random.WithPrefix("consumer"))
		consumerBinding := standardQueueServiceInstance.Bind(consumerApp, services.WithBindingName(consumerBindingName))
		defer consumerBinding.Unbind()

		By("binding the app to the service instance - DLQ consumer")
		bindingDLQName := random.Name(random.WithPrefix("dlq"))
		dlqBinding := dlqServiceInstance.Bind(dlqConsumerApp, services.WithBindingName(bindingDLQName))
		defer dlqBinding.Unbind()

		By("starting the apps")
		apps.Start(producerApp, consumerApp, dlqConsumerApp)

		By("sending a message - producer")
		message := random.Hexadecimal()
		producerApp.POST(message, "/send/%s", producerBindingName)

		By("receiving the message - consumer")
		got := consumerApp.GET("/receive/%s", consumerBindingName).String()
		Expect(got).To(Equal(message))

		By("sending an incorrect message - producer")
		incorrectlyFormattedMessage := "incorrectly formatted"
		producerApp.POST(incorrectlyFormattedMessage, "/send/%s", producerBindingName)

		By("reading an incorrect message - consumer")
		getResponse := consumerApp.GETResponse("/receive_many_messages/%s", consumerBindingName)
		Expect(getResponse).To(HaveHTTPStatus(http.StatusRequestTimeout))
		b, err := io.ReadAll(getResponse.Body)
		Expect(err).ToNot(HaveOccurred(), "error reading response body")
		Expect(string(b)).To(ContainSubstring("context deadline exceeded"), "context deadline exceeded after sending message to the DLQ")
		defer getResponse.Body.Close()

		By("reading message in DLQ - DLQ consumer")
		got = dlqConsumerApp.GET("/dlq/%s", bindingDLQName).String()
		Expect(got).To(Equal(incorrectlyFormattedMessage))

	})
})
