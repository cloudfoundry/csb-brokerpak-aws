package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sqsapp/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func dlqSubscriber(ctx context.Context, creds credentials.Credentials, binding string) (int, string) {

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
				return http.StatusRequestTimeout, "DLQ Subscriber stopped due to context deadline exceeded."
			}

			if ctx.Err() == context.Canceled {
				return http.StatusOK, "DLQ Subscriber stopped due to context cancellation."
			}

			return http.StatusInternalServerError, "DLQ Subscriber stopped due to unknown error."

		default:
			receiveInput := &sqs.ReceiveMessageInput{
				QueueUrl:            &cred.URL,
				MaxNumberOfMessages: 1,
				WaitTimeSeconds:     20,
			}

			var output *sqs.ReceiveMessageOutput
			if output, err = client.ReceiveMessage(ctx, receiveInput); err != nil {
				return http.StatusBadRequest, fmt.Sprintf("error receiving DLQ message: %q", err)
			}

			if len(output.Messages) == 0 {
				log.Print("No messages found in DLQ\n")
				continue
			}

			message := output.Messages[0]
			body := aws.ToString(message.Body)

			log.Printf("DLQ Message logged: %s - binding name: %s\n", body, binding)

			deleteInput := &sqs.DeleteMessageInput{
				QueueUrl:      &cred.URL,
				ReceiptHandle: message.ReceiptHandle,
			}

			if _, err := client.DeleteMessage(ctx, deleteInput); err != nil {
				return http.StatusNotAcceptable, fmt.Sprintf("failed to delete message: %q", err)
			}

			return http.StatusOK, aws.ToString(message.Body)
		}
	}

}
