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
	It("uses a FIFO queue in two apps", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-aws-sqs", services.WithPlan("fifo"))
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

		By("sending a message from producer app")
		message := random.Hexadecimal()
		messageGroupID := random.Hexadecimal()
		messageDeduplicationID := random.Hexadecimal()
		appOne.POST(message, "/send/%s?messageGroupId=%s&messageDeduplicationId=%s", bindingOneName, messageGroupID, messageDeduplicationID)

		By("receiving the message using the consumer app")
		got := appTwo.GET("/retrieve_and_delete/%s", bindingTwoName).String()
		Expect(got).To(Equal(message))
	})

	It("uses a Standard queue with accociated DLQ and triggers redrive", func() {
		By("creating a DLQ service instance")
		dlqServiceInstance := services.CreateInstance(
			"csb-aws-sqs",
			services.WithPlan("standard"),
		)
		defer dlqServiceInstance.Delete()

		csbKey := dlqServiceInstance.CreateServiceKey()
		defer csbKey.Delete()
		var skReceiver struct {
			ARN string `json:"arn"`
		}
		csbKey.Get(&skReceiver)
		dlqARN := skReceiver.ARN

		By("creating a Standard Queue")
		standardQueueServiceInstance := services.CreateInstance(
			"csb-aws-sqs",
			services.WithPlan("standard"),
			services.WithParameters(map[string]any{
				"dlq_arn":           dlqARN,
				"max_receive_count": 1,
			}),
		)
		defer standardQueueServiceInstance.Delete()

		By("pushing the unstarted apps")
		producerApp := apps.Push(apps.WithName(random.Name(random.WithPrefix("producer"))), apps.WithApp(apps.SQS))
		consumerApp := apps.Push(apps.WithName(random.Name(random.WithPrefix("consumer"))), apps.WithApp(apps.SQS))
		defer apps.Delete(producerApp, consumerApp)

		By("binding producer and consumer apps to the standard queue")
		producerBindingName := random.Name(random.WithPrefix("producer"))
		producerBinding := standardQueueServiceInstance.Bind(producerApp, services.WithBindingName(producerBindingName))
		defer producerBinding.Unbind()

		consumerBindingName := random.Name(random.WithPrefix("consumer"))
		consumerBinding := standardQueueServiceInstance.Bind(consumerApp, services.WithBindingName(consumerBindingName))
		defer consumerBinding.Unbind()

		By("binding the consumer app to DLQ")
		bindingDLQName := random.Name(random.WithPrefix("dlq"))
		dlqBinding := dlqServiceInstance.Bind(consumerApp, services.WithBindingName(bindingDLQName))
		defer dlqBinding.Unbind()

		By("starting the apps")
		apps.Start(producerApp, consumerApp)

		By("checking message is send and received between two apps", func() {
			By("sending a message using producer app")
			message := random.Hexadecimal()
			producerApp.POST(message, "/send/%s", producerBindingName)

			By("reading message using consumer app")
			got := consumerApp.GET("/retrieve_and_delete/%s", consumerBindingName).String()
			Expect(got).To(Equal(message))
		})

		By("checking consumer app can read from DLQ", func() {
			By("sending a message using producer app")
			message := random.Hexadecimal()
			producerApp.POST(message, "/send/%s", producerBindingName)

			By("read the message without deleting it using consumer app")
			got := consumerApp.GET("/retrieve/%s", consumerBindingName).String()
			Expect(got).To(Equal(message))

			By("triggering move to DLQ by attempting to retrieve from the queue again using consumer app")
			response := consumerApp.GETResponse("/retrieve/%s", consumerBindingName)
			Expect(response).To(HaveHTTPStatus(http.StatusTooEarly))

			By("reading message in the DLQ using consumer app")
			got = consumerApp.GET("/retrieve_and_delete/%s", bindingDLQName).String()
			Expect(got).To(Equal(message))
		})

		By("checking consumer app can trigger a redrive from DLQ to original queue", func() {
			By("sending another message using producer app")
			message := random.Hexadecimal()
			producerApp.POST(message, "/send/%s", producerBindingName)

			By("read a message without delete it using consumer app")
			got := consumerApp.GET("/retrieve/%s", consumerBindingName).String()
			Expect(got).To(Equal(message))

			By("triggering move to DLQ by attempting to retrieve from the queue again using consumer app")
			response := consumerApp.GETResponse("/retrieve/%s", consumerBindingName)
			Expect(response).To(HaveHTTPStatus(http.StatusTooEarly))

			By("triggering redrive from DLQ to original queue using consumer app")
			consumerApp.POST("", "/redrive/%s?dlq_arn=%s", consumerBindingName, dlqARN)

			By("reading message in the original queue using consumer app")
			got = consumerApp.GET("/retrieve_and_delete/%s", consumerBindingName).String()
			Expect(got).To(Equal(message))
		})
	})
})
