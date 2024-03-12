package app

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"

	"sqsapp/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func handleSend(creds credentials.Credentials) func(r *http.Request) (int, string) {
	return func(r *http.Request) (int, string) {
		binding := r.PathValue("binding_name")
		log.Printf("Handling send on binding %q\n", binding)

		cred, ok := creds[binding]
		if !ok {
			return http.StatusBadRequest, fmt.Sprintf("no creds found for binding: %q", binding)
		}
		cfg, err := cred.Config()
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("could not read AWS config: %q", err)
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			return http.StatusBadRequest, fmt.Sprintf("could not read body: %q", err)
		}
		defer r.Body.Close()
		body := string(data)

		sendMessageInput := &sqs.SendMessageInput{
			MessageBody: &body,
			QueueUrl:    &cred.URL,
		}
		messageGroupID := r.URL.Query().Get("messageGroupId")
		messageDeduplicationID := r.URL.Query().Get("messageDeduplicationId")

		if messageGroupID != "" || messageDeduplicationID != "" {
			sendMessageInput = &sqs.SendMessageInput{
				MessageBody:            &body,
				QueueUrl:               &cred.URL,
				MessageGroupId:         &messageGroupID,
				MessageDeduplicationId: &messageDeduplicationID,
			}
		}

		output, err := sqs.NewFromConfig(cfg).SendMessage(r.Context(), sendMessageInput)
		if err != nil {
			return http.StatusBadRequest, fmt.Sprintf("error sending message: %q", err)
		}

		id := aws.ToString(output.MessageId)
		log.Printf("sent message ID: %q\n", id)
		return http.StatusOK, fmt.Sprintf(`{"id":"%s"}`, id)
	}
}
