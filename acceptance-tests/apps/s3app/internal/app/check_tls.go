package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	appcreds "s3app/internal/credentials"
)

func CheckHTTPSHandler(protocol, slug string, creds appcreds.S3Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Handling request %s.\n", r.URL.Path)

		switch r.Method {
		case http.MethodGet:
			log.Printf("Handling check %s.", protocol)
			file := strings.TrimPrefix(r.URL.Path, slug)

			// Valid URL: https://bucket-name.s3.region-code.amazonaws.com/key-name
			requestURL := fmt.Sprintf("%s://%s.s3.%s.amazonaws.com/%s", protocol, creds.BucketName, creds.Region, file)
			log.Printf("URL: %s", requestURL)
			req, err := http.NewRequest(http.MethodGet, requestURL, nil)
			if err != nil {
				fail(w, http.StatusInternalServerError, "Error creating HTTP request slug %q: %s", file, err)
				return
			}
			client := http.Client{Timeout: 5 * time.Second}
			res, err := client.Do(req)
			if err != nil {
				fail(w, http.StatusInternalServerError, "Error requesting file %q: %s", file, err)
				return
			}

			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				fail(w, res.StatusCode, "Unexpected status code %d - %s", res.StatusCode, res.Status)
				return
			}

			fileContents, err := io.ReadAll(res.Body)
			if err != nil {
				fail(w, http.StatusFailedDependency, "Error reading file %q: %s", file, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "multipart/form-data")
			_, err = w.Write(fileContents)
			if err != nil {
				log.Printf("Error writing value: %s", err)
				return
			}

			log.Printf("File %q is downloaded.", file)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}
