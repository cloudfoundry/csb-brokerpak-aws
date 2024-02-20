package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"sqsapp/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func handleReceiveManyMessages(ctx context.Context, creds credentials.Credentials, binding string) (int, string) {

	log.Printf("Handling receive on binding %q\n", binding)

	cred, ok := creds[binding]
	if !ok {
		return http.StatusBadRequest, fmt.Sprintf("no creds found for binding: %q", binding)
	}

	cfg, err := cred.Config()
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("could not read AWS config: %q", err)
	}

	client := sqs.NewFromConfig(cfg)
	for {
		select {
		case <-ctx.Done():

			if ctx.Err() == context.DeadlineExceeded {
				return http.StatusRequestTimeout, "context deadline exceeded"
			}

			if ctx.Err() == context.Canceled {
				return http.StatusOK, "context cancellation"
			}

			return http.StatusInternalServerError, "unknown error"

		default:
			receiveInput := &sqs.ReceiveMessageInput{
				QueueUrl:            &cred.URL,
				MaxNumberOfMessages: 1,
				WaitTimeSeconds:     10,
				// The duration (in seconds) that the received messages are hidden from subsequent
				// retrieve requests after being retrieved by a ReceiveMessage request.
				VisibilityTimeout: 1,
			}

			var output *sqs.ReceiveMessageOutput
			if output, err = client.ReceiveMessage(ctx, receiveInput); err != nil {
				log.Printf("Error receiving message: %s\n", err)
				continue
			}

			if len(output.Messages) == 0 {
				log.Print("No messages found\n")
				continue
			}

			message := output.Messages[0]
			body := aws.ToString(message.Body)

			log.Printf("Message received: %s - binding name: %s\n", body, binding)

			// Hack to simulate a message that is incorrectly formatted
			if strings.Contains(body, "incorrectly formatted") {
				log.Printf("incorrectly formatted message receive on binding %q, leaving in queue for DLQ\n", binding)
				continue
			}

			deleteInput := &sqs.DeleteMessageInput{
				QueueUrl:      &cred.URL,
				ReceiptHandle: message.ReceiptHandle,
			}

			if _, err := client.DeleteMessage(ctx, deleteInput); err != nil {
				return http.StatusNotAcceptable, fmt.Sprintf("failed to delete message: %q", err)
			}

			log.Printf("Message %q received.\n", aws.ToString(message.Body))
			return http.StatusOK, aws.ToString(message.Body)
		}
	}
}
