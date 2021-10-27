package app

import (
	"log"
	"net/http"
	"s3app/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
)

func handleDelete(client *s3.Client, creds credentials.S3Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling delete.")

		filename, ok := mux.Vars(r)["file_name"]
		if !ok {
			fail(w, http.StatusBadRequest, "File name missing")
			return
		}

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
}
