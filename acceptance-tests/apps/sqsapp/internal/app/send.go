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

func handleSend(creds credentials.Credentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling send.")

		data, err := io.ReadAll(r.Body)
		if err != nil {
			fail(w, http.StatusBadRequest, "could not read body: %q", err)
			return
		}
		defer r.Body.Close()
		body := string(data)

		cfg, err := creds.Config()
		if err != nil {
			fail(w, http.StatusInternalServerError, "could not read AWS config: %q", err)
			return
		}

		output, err := sqs.NewFromConfig(cfg).SendMessage(r.Context(), &sqs.SendMessageInput{
			MessageBody: &body,
			QueueUrl:    &creds.URL,
		})
		if err != nil {
			fail(w, http.StatusBadRequest, "error sending message: %q", err)
			return
		}

		id := aws.ToString(output.MessageId)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"id":"%s"}`, id)))

		log.Printf("sent message ID: %q\n", id)
	}
}
