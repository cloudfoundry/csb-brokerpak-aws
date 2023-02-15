package app

import (
	"context"
	appcreds "dynamodbapp/internal/credentials"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func App(creds appcreds.DynamoDBService) http.HandlerFunc {
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

	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodGet:
			handleGet(w, r, key, client, creds)
		case http.MethodPut:
			handleSet(w, r, key, client, creds)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
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
