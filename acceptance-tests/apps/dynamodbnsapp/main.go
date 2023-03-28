package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"dynamodbnsapp/internal/app"
	"dynamodbnsapp/internal/credentials"
)

func main() {
	log.Println("Starting.")

	log.Println("Reading credentials.")
	creds, err := credentials.Read()
	if err != nil {
		panic(err)
	}

	port := appPort()
	log.Printf("Listening on port: %s", port)
	appRouter := app.App(creds)
	_ = http.ListenAndServe(port, appRouter)
}

func appPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
