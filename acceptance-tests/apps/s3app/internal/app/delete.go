package app

import (
	"log"
	"net/http"
	"s3app/internal/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
)

func handleDelete(session *session.Session, creds credentials.S3Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling delete.")

		filename, ok := mux.Vars(r)["file_name"]
		if !ok {
			log.Println("File name missing.")
			http.Error(w, "File name missing.", http.StatusBadRequest)
			return
		}

		svc := s3.New(session)
		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(creds.BucketName),
			Delete: &s3.Delete{
				Objects: []*s3.ObjectIdentifier{
					{
						Key: aws.String(filename),
					},
				},
				Quiet: aws.Bool(true),
			},
		}
		_, err := svc.DeleteObjects(input)
		if err != nil {
			log.Printf("Error deleting file %q: %s", filename, err)
			http.Error(w, "Failed to delete file.", http.StatusFailedDependency)
			return
		}

		w.WriteHeader(http.StatusGone)
		log.Printf("File %q is deleted.", filename)
	}
}
