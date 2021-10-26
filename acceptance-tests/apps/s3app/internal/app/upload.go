package app

import (
	"log"
	"net/http"
	"s3app/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
)

func handleUpload(client *s3.Client, creds credentials.S3Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling upload.")

		filename, ok := mux.Vars(r)["file_name"]
		if !ok {
			log.Println("File name missing.")
			http.Error(w, "File name missing.", http.StatusBadRequest)
			return
		}

		_, err := client.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket:        aws.String(creds.BucketName),
			Key:           aws.String(filename),
			Body:          r.Body,
			ContentLength: r.ContentLength,
		})
		if err != nil {
			fail(w, http.StatusFailedDependency, "Error uploading file part %q: %s", filename, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("File %q is uploaded to bucket %q.", filename, creds.BucketName)
	}
}
