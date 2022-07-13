package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	appcreds "s3app/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
)

func App(creds appcreds.S3Service) *mux.Router {
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

	r := mux.NewRouter()

	r.HandleFunc("/", aliveness).Methods(http.MethodHead, http.MethodGet)
	r.HandleFunc("/{file_name}", handleUpload(client, creds)).Methods(http.MethodPut)
	r.HandleFunc("/{file_name}", handleDownload(client, creds)).Methods(http.MethodGet)
	r.HandleFunc("/{file_name}", handleDelete(client, creds)).Methods(http.MethodDelete)

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
