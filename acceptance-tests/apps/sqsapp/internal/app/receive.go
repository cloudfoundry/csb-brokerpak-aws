package app

import (
	"log"
	"net/http"
	"sqsapp/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func handleReceive(creds credentials.Credentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling receive.")

		cfg, err := creds.Config()
		if err != nil {
			fail(w, http.StatusInternalServerError, "could not read AWS config: %q", err)
			return
		}

		client := sqs.NewFromConfig(cfg)
		output, err := client.ReceiveMessage(r.Context(), &sqs.ReceiveMessageInput{
			QueueUrl:          &creds.URL,
			VisibilityTimeout: 5, // we delete the message immediately below
			WaitTimeSeconds:   20,
		})
		switch {
		case err != nil:
			fail(w, http.StatusBadRequest, "error receiving message: %q", err)
			return
		case len(output.Messages) == 0:
			fail(w, http.StatusTooEarly, "no messages received")
			return
		}

		message := output.Messages[0]
		_, err = client.DeleteMessage(r.Context(), &sqs.DeleteMessageInput{
			QueueUrl:      &creds.URL,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			fail(w, http.StatusNotAcceptable, "failed to delete message: %q", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(*message.Body))

		log.Printf("Message %q received.\n", *message.Body)
	}
}
