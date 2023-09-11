package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/endpointcreds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	appcreds "dynamodbtableapp/internal/credentials"
)

func App(creds appcreds.DynamoDBService) http.HandlerFunc {
	authToken := base64.StdEncoding.EncodeToString([]byte(strings.ReplaceAll(creds.AccessKeyId, ":", "*") + ":" + creds.AccessKeySecret))
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(
				endpointcreds.New(creds.CredsEndpoint, func(options *endpointcreds.Options) {
					options.AuthorizationToken = "Basic " + authToken
				}),
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

func aliveness(w http.ResponseWriter, _ *http.Request) {
	log.Printf("Handled aliveness test.")
	w.WriteHeader(http.StatusNoContent)
}

func fail(w http.ResponseWriter, code int, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Println(msg)
	http.Error(w, msg, code)
}
