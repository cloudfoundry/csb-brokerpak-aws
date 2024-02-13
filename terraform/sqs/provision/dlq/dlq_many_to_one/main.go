package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var (
	myQueueURL                           = "MY_QUEUE_URL"
	myQueueTwoURL                        = "MY_QUEUE_TWO_URL"
	myDlqURL                             = "MY_DLQ_URL"
	awsUserAccessKeyIdStandardQueues     = "USER_ACCESS_KEY_ID_STANDARD_QUEUES"
	awsUserSecretAccessKeyStandardQueues = "USER_SECRET_ACCESS_KEY_STANDARD_QUEUES"
	awsUserAccessKeyIDDLQ                = "USER_ACCESS_KEY_ID_DLQ"
	awsUserSecretAccessKeyDLQ            = "USER_SECRET_ACCESS_KEY_DLQ"
)

func main() {
	var envs = map[string]string{
		myQueueURL:                           "",
		myQueueTwoURL:                        "",
		myDlqURL:                             "",
		awsUserAccessKeyIdStandardQueues:     "",
		awsUserSecretAccessKeyStandardQueues: "",
		awsUserAccessKeyIDDLQ:                "",
		awsUserSecretAccessKeyDLQ:            "",
	}

	for k := range envs {
		envs[k] = os.Getenv(k)
		if envs[k] == "" {
			log.Fatalf("environment variable: %s must be set", k)
		}
	}

	cfgStandardQueues, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(
				credentials.NewStaticCredentialsProvider(
					envs[awsUserAccessKeyIdStandardQueues],
					envs[awsUserSecretAccessKeyStandardQueues],
					"",
				),
			),
		),
		config.WithRegion("us-west-2"),
	)

	if err != nil {
		log.Fatalf("invalid AWS configuration %s", err.Error())
	}

	cfgDLQ, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(
				credentials.NewStaticCredentialsProvider(
					envs[awsUserAccessKeyIDDLQ],
					envs[awsUserSecretAccessKeyDLQ],
					"",
				),
			),
		),
		config.WithRegion("us-west-2"),
	)

	if err != nil {
		log.Fatalf("invalid AWS configuration for DLQ %s", err.Error())
	}

	svc := sqs.NewFromConfig(cfgStandardQueues)
	svcDLQ := sqs.NewFromConfig(cfgDLQ)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// queue 1
	go producer(ctx, svc, envs[myQueueURL], "queue_one")
	go subscriber(ctx, svc, envs[myQueueURL])

	// queue 2
	go producer(ctx, svc, envs[myQueueTwoURL], "queue_two")
	go subscriber(ctx, svc, envs[myQueueTwoURL])

	// Same DLQ
	go dlqSubscriber(ctx, svcDLQ, envs[myDlqURL])

	<-ctx.Done()
	switch ctx.Err() {
	case context.DeadlineExceeded:
		log.Println("Application stopped due to timeout.")
	case context.Canceled:
		log.Println("Application stopped.")
	default:
		log.Println("Application stopped unexpectedly.")
	}
}

func producer(ctx context.Context, svc *sqs.Client, queueURL, queueName string) {
	messages := []string{"This is a correctly formatted message.", "Incorrect format message"}

	for _, messageBody := range messages {
		select {
		case <-ctx.Done():
			log.Println("Producer stopped due to context cancellation.")
			return

		default:
			_, err := svc.SendMessage(ctx, &sqs.SendMessageInput{
				MessageBody: aws.String(messageBody),
				QueueUrl:    &queueURL,
				MessageAttributes: map[string]types.MessageAttributeValue{
					"Queue": {DataType: aws.String("String"), StringValue: aws.String(queueName)},
				},
			})
			if err != nil {
				log.Printf("Error sending message: %s\n", err)
				continue
			}

			log.Printf("Message sent: %s\n", messageBody)
			time.Sleep(1 * time.Second) // Simulate delay between messages
		}
	}
}

func subscriber(ctx context.Context, svc *sqs.Client, queueURL string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Subscriber stopped due to context cancellation.")
			return

		default:
			result, err := svc.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            &queueURL,
				MaxNumberOfMessages: 1,
				WaitTimeSeconds:     10,
				// The duration (in seconds) that the received messages are hidden from subsequent
				// retrieve requests after being retrieved by a ReceiveMessage request.
				VisibilityTimeout:     1,
				MessageAttributeNames: []string{"All"},
			})
			if err != nil {
				log.Printf("Error receiving message: %s\n", err)
				continue
			}

			if len(result.Messages) == 0 {
				log.Print("No messages found\n")
				continue
			}

			body := aws.ToString(result.Messages[0].Body)
			queue := aws.ToString(result.Messages[0].MessageAttributes["Queue"].StringValue)
			log.Printf("Message received: %s - queue %s\n", body, queue)
			if strings.Contains(body, "correctly formatted") {
				_, err = svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      &queueURL,
					ReceiptHandle: result.Messages[0].ReceiptHandle,
				})
				if err != nil {
					log.Printf("Error deleting message: %s\n", err)
					continue
				}

				log.Println("Message processed and deleted.")

			} else {
				log.Println("Message format incorrect, leaving in queue for DLQ")
				// Simulate processing failure
			}
		}
	}
}

func dlqSubscriber(ctx context.Context, svc *sqs.Client, dlqURL string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("DLQ Subscriber stopped due to context cancellation.")
			return
		default:
			result, err := svc.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:              &dlqURL,
				MaxNumberOfMessages:   1,
				WaitTimeSeconds:       20,
				MessageAttributeNames: []string{"All"},
			})
			if err != nil {
				log.Printf("Error receiving DLQ message: %s\n", err)
				continue
			}
			if len(result.Messages) == 0 {
				log.Print("No messages found in DLQ\n")
				continue
			}

			body := aws.ToString(result.Messages[0].Body)
			queue := aws.ToString(result.Messages[0].MessageAttributes["Queue"].StringValue)
			log.Printf("DLQ Message logged: %s - queue name: %s\n", body, queue)
		}
	}
}
