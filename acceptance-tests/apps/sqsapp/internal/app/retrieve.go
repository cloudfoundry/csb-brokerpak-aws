package app

import (
	"fmt"
	"log"
	"net/http"

	"sqsapp/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func handleRetrieve(creds credentials.Credentials) func(r *http.Request) (int, string) {
	return func(r *http.Request) (int, string) {
		binding := r.PathValue("binding_name")
		log.Printf("Handling retrieve on binding %q\n", binding)

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
			QueueUrl: &cred.URL,
			// The duration (in seconds) that the received messages are hidden from subsequent
			// retrieve requests after being retrieved by a ReceiveMessage request.
			VisibilityTimeout: 1,
			WaitTimeSeconds:   10,
		})
		switch {
		case err != nil:
			return http.StatusBadRequest, fmt.Sprintf("error receiving message: %q", err)
		case len(output.Messages) == 0:
			return http.StatusTooEarly, "no messages received"
		}

		message := output.Messages[0]
		log.Printf("Message %q received.\n", aws.ToString(message.Body))
		return http.StatusOK, aws.ToString(message.Body)
	}
}
