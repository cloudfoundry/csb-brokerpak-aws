package app

import (
	"dynamodbapp/internal/credentials"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func handleSet(client *dynamodb.Client, creds credentials.DynamoDBService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling set.")

		key, ok := mux.Vars(r)["key"]
		if !ok {
			log.Println("Key missing.")
			fail(w, http.StatusBadRequest, "Key missing.")
			return
		}

		value, err := io.ReadAll(r.Body)
		if err != nil {
			fail(w, http.StatusBadRequest, "Error parsing value: %s", err)
			return
		}

		_, err = client.PutItem(r.Context(), &dynamodb.PutItemInput{
			TableName: aws.String(creds.TableName),
			Item: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: uuid.New()},
				"key":   &types.AttributeValueMemberS{Value: key},
				"value": &types.AttributeValueMemberS{Value: string(value)},
			},
		})
		if err != nil {
			fail(w, http.StatusFailedDependency, "Error creating item with key %q and value %q in table %q: %s", key, value, creds.TableName, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		log.Printf("Created item with key %q and value %q in table %q.", key, value, creds.TableName)
	}
}
