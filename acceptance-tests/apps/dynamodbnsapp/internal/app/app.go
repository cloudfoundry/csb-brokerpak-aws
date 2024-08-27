package app

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	appcreds "dynamodbnsapp/internal/credentials"
)

const (
	ddbClientKey = "ddbClient"
)

func App(creds appcreds.DynamoDBNamespaceService) http.Handler {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(
				credentials.NewStaticCredentialsProvider(
					creds.AccessKeyID,
					creds.SecretAccessKey,
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

	r := http.NewServeMux()

	r.HandleFunc("POST /tables", ddbClient(client, createTable))
	r.HandleFunc("GET /tables/{tableName}", ddbClient(client, tableCtx(getTable)))
	r.HandleFunc("DELETE /tables/{tableName}", ddbClient(client, tableCtx(deleteTable)))
	r.HandleFunc("POST /tables/{tableName}/values/{key}", ddbClient(client, tableCtx(tableKeyCtx(createValue))))
	r.HandleFunc("GET /tables/{tableName}/values/{key}/{pk}", ddbClient(client, tableCtx(tableKeyCtx(tablePrimaryKeyCtx(getValue)))))
	r.HandleFunc("DELETE /tables/{tableName}/values/{key}/{pk}", ddbClient(client, tableCtx(tableKeyCtx(tablePrimaryKeyCtx(deleteValue)))))

	return r
}

func ddbClient(client *dynamodb.Client, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ddbClientKey, client)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
