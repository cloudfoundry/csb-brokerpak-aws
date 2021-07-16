package app

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"s3app/internal/credentials"
)

func handleDownload(session *session.Session, cred *credentials.S3Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling download.")

		filename, ok := mux.Vars(r)["file_name"]
		if !ok {
			log.Println("File name missing.")
			http.Error(w, "File name missing.", http.StatusBadRequest)
			return
		}

		file, err := os.Create(filename)
		downloader := s3manager.NewDownloader(session)
		_, err = downloader.Download(file, &s3.GetObjectInput{
			Bucket: aws.String(cred.BucketName),
			Key:    aws.String(filename),
		})
		if err != nil {
			log.Printf("Error dowloading file %q: %s", filename, err)
			http.Error(w, "Failed to download file.", http.StatusFailedDependency)
			return
		}

		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			log.Printf("Error reading file %q: %s", filename, err)
			http.Error(w, "Failed to read file.", http.StatusFailedDependency)
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
}
