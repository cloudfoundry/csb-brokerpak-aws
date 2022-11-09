package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"postgresqlapp/internal/app"
	"postgresqlapp/internal/connector"
)

func main() {
	log.Println("Starting.")
	conn, err := connector.New()
	if err != nil {
		panic(err)
	}

	port := port()
	log.Printf("Listening on port: %s", port)
	http.Handle("/", app.App(conn))
	_ = http.ListenAndServe(port, nil)
}

func port() string {
	if port := os.Getenv("PORT"); port != "" {
		return fmt.Sprintf(":%s", port)
	}
	return ":8080"
}
