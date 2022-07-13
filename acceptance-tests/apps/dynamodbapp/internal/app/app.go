package app

import (
	"context"
	appcreds "dynamodbapp/internal/credentials"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gorilla/mux"
)

func App(creds appcreds.DynamoDBService) *mux.Router {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(
				credentials.NewStaticCredentialsProvider(
					creds.AccessKeyId,
					creds.AccessKeySecret,
					"",
				),
			),
		),
		config.WithRegion(creds.Region),
	)
	if err != nil {
		panic(err)
	}

	client := dynamodb.NewFromConfig(cfg)
	r := mux.NewRouter()

	r.HandleFunc("/", aliveness).Methods(http.MethodHead, http.MethodGet)
	r.HandleFunc("/{key}", handleSet(client, creds)).Methods(http.MethodPut)
	r.HandleFunc("/{key}", handleGet(client, creds)).Methods(http.MethodGet)

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
