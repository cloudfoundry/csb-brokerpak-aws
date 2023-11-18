package app

import (
	"log"
	"net/http"
	"path"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"s3app/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func HandleUploadWithACL(client *credentials.Client, acl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling upload with ACL.")
		switch r.Method {
		case http.MethodPut:
			filename := path.Base(r.URL.Path)

			_, err := client.S3Client.PutObject(r.Context(), &s3.PutObjectInput{
				Bucket:        aws.String(client.Credentials.BucketName),
				Key:           aws.String(filename),
				Body:          r.Body,
				ContentLength: &r.ContentLength,
				ACL:           types.ObjectCannedACL(acl),
			})
			if err != nil {
				fail(w, http.StatusFailedDependency, "Error uploading file part %q: %s", filename, err)
				return
			}

			w.WriteHeader(http.StatusCreated)
			log.Printf("File %q is uploaded to bucket %q.", filename, client.Credentials.BucketName)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}
