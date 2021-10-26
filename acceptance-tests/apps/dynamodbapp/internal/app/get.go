package app

import (
	"dynamodbapp/internal/credentials"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gorilla/mux"
)

func handleGet(client *dynamodb.Client, creds credentials.DynamoDBService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling get.")

		key, ok := mux.Vars(r)["key"]
		if !ok {
			fail(w, http.StatusBadRequest, "Key missing.")
			return
		}

		out, err := client.Scan(r.Context(), &dynamodb.ScanInput{
			TableName: aws.String(creds.TableName),
		})
		if err != nil {
			fail(w, http.StatusNotFound, "failed to scan table %q: %s", creds.TableName, err)
			return
		}

		var value string
		for _, i := range out.Items {
			if k, ok := i["key"].(*types.AttributeValueMemberS); ok && k.Value == key {
				if v, ok := i["value"].(*types.AttributeValueMemberS); ok {
					value = v.Value
				}
			}
		}

		if value == "" {
			fail(w, http.StatusNotFound, "failed to find item with key %q", key)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write([]byte(value))

		log.Printf("Value %q retrived from item with key %q in table %q.", value, key, creds.TableName)
	}
}
