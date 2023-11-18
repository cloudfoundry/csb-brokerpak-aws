package app

import (
	"log"
	"net/http"

	"s3app/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handleUpload(w http.ResponseWriter, r *http.Request, filename string, client *credentials.Client) {
	log.Println("Handling upload.")

	_, err := client.S3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:        aws.String(client.Credentials.BucketName),
		Key:           aws.String(filename),
		Body:          r.Body,
		ContentLength: &r.ContentLength,
	})
	if err != nil {
		fail(w, http.StatusFailedDependency, "Error uploading file part %q: %s", filename, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("File %q is uploaded to bucket %q.", filename, client.Credentials.BucketName)
}
