package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"s3app/internal/app"
	"s3app/internal/credentials"
)

func main() {
	log.Println("Starting.")

	log.Println("Reading credentials.")
	client := credentials.NewClient()

	p := port()
	log.Printf("Listening on port: %s", p)

	http.Handle("/", app.App(client))
	http.Handle("/check-https/", app.CheckHTTPSHandler("https", client.Credentials))
	http.Handle("/check-http/", app.CheckHTTPSHandler("http", client.Credentials))
	http.Handle("/upload-with-public-read-acl/", app.HandleUploadWithACL(client, "public-read"))
	if err := http.ListenAndServe(p, nil); err != http.ErrServerClosed {
		panic(err)
	}
}

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
