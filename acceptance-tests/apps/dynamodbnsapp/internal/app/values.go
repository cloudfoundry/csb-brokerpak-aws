package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	tableKeyNameKey        = "key"
	tablePrimaryKeyNameKey = "pk"
)

var (
	itemNotFoundErr = errors.New("item not found")
)

type ValueResponse struct {
	Pk      int64  `json:"pk"`
	Sorting string `json:"sorting"`
	Value   string `json:"value"`
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
		writeJSONResponse(w, statusCodeFromAWSError(err), NewErrResponse(err))
		return
	}

	if reply.Attributes == nil {
		writeJSONResponse(w, http.StatusNotFound, NewErrResponse(itemNotFoundErr))
		return
	}

	pk, _ := strconv.ParseInt(reply.Attributes[tableKeyPrimary].(*types.AttributeValueMemberN).Value, 10, 32)
	writeJSONResponse(w, http.StatusOK, ValueResponse{
		Pk:      pk,
		Sorting: reply.Attributes[tableKeySorting].(*types.AttributeValueMemberS).Value,
		Value:   reply.Attributes[tableValueColumnName].(*types.AttributeValueMemberS).Value,
	})
}

func createValue(w http.ResponseWriter, r *http.Request) {
	value, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONResponse(w, http.StatusBadRequest, NewErrResponse(err))
		return
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
		writeJSONResponse(w, statusCodeFromAWSError(err), NewErrResponse(err))
		return
	}

	writeJSONResponse(w, http.StatusCreated, ValueResponse{
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
		writeJSONResponse(w, statusCodeFromAWSError(err), NewErrResponse(err))
		return
	}

	if reply.Item == nil {
		writeJSONResponse(w, http.StatusNotFound, NewErrResponse(itemNotFoundErr))
		return
	}

	pk, err := strconv.ParseInt(reply.Item[tableKeyPrimary].(*types.AttributeValueMemberN).Value, 10, 64)
	if err != nil {
		writeJSONResponse(w, http.StatusUnprocessableEntity, NewErrResponse(err))
		return
	}

	writeJSONResponse(w, http.StatusOK, ValueResponse{
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
