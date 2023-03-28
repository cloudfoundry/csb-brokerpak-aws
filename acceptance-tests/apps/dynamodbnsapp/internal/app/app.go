package app

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	appcreds "dynamodbnsapp/internal/credentials"
)

const (
	ddbClientKey = "ddbClient"
)

func App(creds appcreds.DynamoDBNamespaceService) chi.Router {
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

	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.With(ddbClient(client)).Route("/tables", func(rt chi.Router) {
		rt.Post("/", createTable)
		rt.With(tableCtx).Get("/{tableName}", getTable)
		rt.With(tableCtx).Delete("/{tableName}", deleteTable)
		rt.With(tableCtx, tableKeyCtx).Route("/{tableName}/values/{key}", func(rv chi.Router) {
			rv.Post("/", createValue)
			rv.With(tablePrimaryKeyCtx).Route("/{pk}", func(rp chi.Router) {
				rp.Get("/", getValue)
				rp.Delete("/", deleteValue)
			})
		})
	})

	return r
}

func ddbClient(client *dynamodb.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ddbClientKey, client)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
