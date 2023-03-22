package app

import (
	"dynamodbtableapp/internal/credentials"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pborman/uuid"
)

func handleSet(w http.ResponseWriter, r *http.Request, key string, client *dynamodb.Client, creds credentials.DynamoDBService) {
	log.Println("Handling set.")

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
