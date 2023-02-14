package app

import (
	"log"
	"net/http"
	"s3app/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handleUpload(w http.ResponseWriter, r *http.Request, filename string, client *s3.Client, creds credentials.S3Service) {
	log.Println("Handling upload.")

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
