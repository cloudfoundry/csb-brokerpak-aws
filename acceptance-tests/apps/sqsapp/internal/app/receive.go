package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"

	"sqsapp/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func handleReceive(creds credentials.Credentials) func(r *http.Request) (int, string) {
	return func(r *http.Request) (int, string) {
		binding := r.PathValue("binding_name")
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
		output, err := client.ReceiveMessage(r.Context(), &sqs.ReceiveMessageInput{
			QueueUrl:          &cred.URL,
			VisibilityTimeout: 5, // we delete the message immediately below
			WaitTimeSeconds:   20,
		})
		switch {
		case err != nil:
			return http.StatusBadRequest, fmt.Sprintf("error receiving message: %q", err)
		case len(output.Messages) == 0:
			return http.StatusTooEarly, "no messages received"
		}

		message := output.Messages[0]
		_, err = client.DeleteMessage(r.Context(), &sqs.DeleteMessageInput{
			QueueUrl:      &cred.URL,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			return http.StatusNotAcceptable, fmt.Sprintf("failed to delete message: %q", err)
		}

		log.Printf("Message %q received.\n", aws.ToString(message.Body))
		return http.StatusOK, aws.ToString(message.Body)
	}
}
