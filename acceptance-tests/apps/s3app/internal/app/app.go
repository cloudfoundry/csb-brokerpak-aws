package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	appcreds "s3app/internal/credentials"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func App(creds appcreds.S3Service) http.HandlerFunc {
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

	client := s3.NewFromConfig(cfg)

	return func(w http.ResponseWriter, r *http.Request) {
		filename := strings.Trim(r.URL.Path, "/")
		switch r.Method {
		case http.MethodHead:
			aliveness(w, r)
		case http.MethodPut:
			handleUpload(w, r, filename, client, creds)
		case http.MethodGet:
			handleDownload(w, r, filename, client, creds)
		case http.MethodDelete:
			handleDelete(w, r, filename, client, creds)
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
