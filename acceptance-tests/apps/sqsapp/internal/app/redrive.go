package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"

	"sqsapp/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func handleRedrive(creds credentials.Credentials) func(r *http.Request) (int, string) {
	return func(r *http.Request) (int, string) {
		binding := r.PathValue("binding_name")
		log.Printf("Handling redrive on binding %q\n", binding)

		cred, ok := creds[binding]
		if !ok {
			return http.StatusBadRequest, fmt.Sprintf("no creds found for binding: %q", binding)
		}
		cfg, err := cred.Config()
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("could not read AWS config: %q", err)
		}

		dlqARN := r.URL.Query().Get("dlq_arn")
		if dlqARN == "" {
			return http.StatusBadRequest, "Should include dlq_arn query param."
		}

		client := sqs.NewFromConfig(cfg)
		output, err := client.StartMessageMoveTask(r.Context(), &sqs.StartMessageMoveTaskInput{
			SourceArn:      &dlqARN,
			DestinationArn: &cred.ARN,
		})
		if err != nil {
			return http.StatusBadRequest, fmt.Sprintf("error starting message move task: %q", err)
		}

		id := aws.ToString(output.TaskHandle)
		log.Printf("started task ID: %q\n", id)
		return http.StatusOK, fmt.Sprintf(`{"id":"%s"}`, id)
	}
}
