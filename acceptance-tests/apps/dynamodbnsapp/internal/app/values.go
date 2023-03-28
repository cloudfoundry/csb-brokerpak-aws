package app

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-chi/render"
)

const (
	tableKeyNameKey        = "key"
	tablePrimaryKeyNameKey = "pk"
)

type ValueResponse struct {
	Pk      int64  `json:"pk"`
	Sorting string `json:"sorting"`
	Value   string `json:"value"`

	RenderableResponse
}

func deleteValue(w http.ResponseWriter, r *http.Request) {
	tableName, key, requestPk, client := extractValueContextValues(r)

	reply, err := client.DeleteItem(r.Context(), &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			tableKeyPrimary: &types.AttributeValueMemberN{Value: requestPk},
			tableKeySorting: &types.AttributeValueMemberS{Value: key},
		},
		TableName:    &tableName,
		ReturnValues: types.ReturnValueAllOld,
	})

	if err != nil {
		_ = render.Render(w, r, errorResponseFromAWSError(err))
		return
	}

	if reply.Attributes == nil {
		_ = render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, StatusText: "Item not found"})
		return
	}

	pk, _ := strconv.ParseInt(reply.Attributes[tableKeyPrimary].(*types.AttributeValueMemberN).Value, 10, 32)
	_ = render.Render(w, r, &ValueResponse{
		Pk:      pk,
		Sorting: reply.Attributes[tableKeySorting].(*types.AttributeValueMemberS).Value,
		Value:   reply.Attributes[tableValueColumnName].(*types.AttributeValueMemberS).Value,
	})
}

func createValue(w http.ResponseWriter, r *http.Request) {
	value, err := io.ReadAll(r.Body)
	if err != nil {
		_ = render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: err.Error()})
	}

	tableName, key, _, client := extractValueContextValues(r)

	primaryKeyValue := time.Now().Unix()
	_, err = client.PutItem(r.Context(), &dynamodb.PutItemInput{
		TableName: &tableName,
		Item: map[string]types.AttributeValue{
			tableKeyPrimary:      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", primaryKeyValue)},
			tableKeySorting:      &types.AttributeValueMemberS{Value: key},
			tableValueColumnName: &types.AttributeValueMemberS{Value: string(value)},
		},
	})

	if err != nil {
		_ = render.Render(w, r, errorResponseFromAWSError(err))
		return
	}

	render.Status(r, http.StatusCreated)
	_ = render.Render(w, r, &ValueResponse{
		Pk:      primaryKeyValue,
		Sorting: key,
		Value:   string(value),
	})
}

func getValue(w http.ResponseWriter, r *http.Request) {
	tableName, key, requestPk, client := extractValueContextValues(r)

	reply, err := client.GetItem(r.Context(), &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			tableKeyPrimary: &types.AttributeValueMemberN{Value: requestPk},
			tableKeySorting: &types.AttributeValueMemberS{Value: key},
		},
		TableName: &tableName,
	})

	if err != nil {
		_ = render.Render(w, r, errorResponseFromAWSError(err))
		return
	}

	if reply.Item == nil {
		_ = render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusNotFound, StatusText: "Item not found"})
		return
	}

	var pk int64
	pk, err = strconv.ParseInt(reply.Item[tableKeyPrimary].(*types.AttributeValueMemberN).Value, 10, 64)
	if err != nil {
		_ = render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusUnprocessableEntity, Err: err})
	}

	_ = render.Render(w, r, &ValueResponse{
		Pk:      pk,
		Sorting: reply.Item[tableKeySorting].(*types.AttributeValueMemberS).Value,
		Value:   reply.Item[tableValueColumnName].(*types.AttributeValueMemberS).Value,
	})
}

func extractValueContextValues(r *http.Request) (tableName string, key string, pk string, client *dynamodb.Client) {
	tableName, client = extractTableContextValues(r)
	key = r.Context().Value(tableKeyNameKey).(string)
	if r.Context().Value(tablePrimaryKeyNameKey) != nil {
		pk = r.Context().Value(tablePrimaryKeyNameKey).(string)
	}
	return tableName, key, pk, client
}
