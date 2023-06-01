package app

import (
	"io"
	"log"
	"net/http"

	"s3app/internal/credentials"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handleDownload(w http.ResponseWriter, r *http.Request, filename string, client *credentials.Client) {
	log.Println("Handling download.")

	obj, err := client.S3Client.GetObject(r.Context(), &s3.GetObjectInput{
		Bucket: aws.String(client.Credentials.BucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		fail(w, http.StatusFailedDependency, "Error downloading file %q from bucket %q: %s", filename, client.Credentials.BucketName, err)
		return
	}

	fileContents, err := io.ReadAll(obj.Body)
	if err != nil {
		fail(w, http.StatusFailedDependency, "Error reading file %q: %s", filename, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "multipart/form-data")
	_, err = w.Write(fileContents)
	if err != nil {
		log.Printf("Error writing value: %s", err)
		return
	}

	log.Printf("File %q is downloaded.", filename)
}
