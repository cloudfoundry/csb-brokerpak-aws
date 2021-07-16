package app

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"net/http"
	"s3app/internal/credentials"

	"github.com/gorilla/mux"
)

func handleUpload(session *session.Session, cred *credentials.S3Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling upload.")

		filename, ok := mux.Vars(r)["file_name"]
		if !ok {
			log.Println("File name missing.")
			http.Error(w, "File name missing.", http.StatusBadRequest)
			return
		}

		uploader := s3manager.NewUploader(session)
		_, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(cred.BucketName),
			Key:    aws.String(filename),
			Body:   r.Body,
		})
		if err != nil {
			log.Printf("Error uploading file %q: %s", filename, err)
			http.Error(w, "Failed to upload file.", http.StatusFailedDependency)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("File %q is uploaded to bucket %q.", filename, cred.BucketName)
	}
}
