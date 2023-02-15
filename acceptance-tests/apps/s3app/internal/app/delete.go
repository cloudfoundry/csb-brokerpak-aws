package app

import (
	"log"
	"net/http"
	"s3app/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handleDelete(w http.ResponseWriter, r *http.Request, filename string, client *s3.Client, creds credentials.S3Service) {
	log.Println("Handling delete.")

	input := s3.DeleteObjectInput{
		Bucket: aws.String(creds.BucketName),
		Key:    aws.String(filename),
	}
	if _, err := client.DeleteObject(r.Context(), &input); err != nil {
		fail(w, http.StatusFailedDependency, "Error deleting file %q: %s", filename, err)
		return
	}

	w.WriteHeader(http.StatusGone)
	log.Printf("File %q is deleted.", filename)
}
