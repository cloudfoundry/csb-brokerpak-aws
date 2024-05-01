package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableNameKey = "tableName"

var (
	tableKeyPrimary      = "pk"
	tableKeySorting      = "sorting"
	tableValueColumnName = "value"
)

type CreateTableRequest struct {
	TableName string `json:"table_name"`
}

type TableCreateResponse struct {
	TableName  string `json:"table_name"`
	CreatedAt  string `json:"created_at"`
	TableClass string `json:"table_class"`
}

type TableDeleteResponse struct {
	TableName string `json:"table_name"`
}

type TableGetResponse struct {
	TableName string `json:"table_name"`
}

func getTable(w http.ResponseWriter, r *http.Request) {
	tableName, client := extractTableContextValues(r)

	_, err := client.DescribeTable(r.Context(), &dynamodb.DescribeTableInput{TableName: &tableName})
	if err != nil {
		writeJSONResponse(w, statusCodeFromAWSError(err), NewErrResponse(err))
		return
	}

	writeJSONResponse(w, http.StatusOK, TableGetResponse{TableName: tableName})
}

func createTable(w http.ResponseWriter, r *http.Request) {
	client := r.Context().Value(ddbClientKey).(*dynamodb.Client)
	var createTableRequest CreateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&createTableRequest); err != nil {
		writeJSONResponse(w, http.StatusBadRequest, NewErrResponse(err))
		return
	}

	params := tableStructureParams(createTableRequest.TableName)
	table, err := client.CreateTable(r.Context(), &params)
	if err != nil {
		writeJSONResponse(w, statusCodeFromAWSError(err), NewErrResponse(err))
		return
	}
	response := TableCreateResponse{
		TableName:  *table.TableDescription.TableName,
		CreatedAt:  table.TableDescription.CreationDateTime.String(),
		TableClass: string(table.TableDescription.TableClassSummary.TableClass),
	}
	writeJSONResponse(w, http.StatusAccepted, response)
}

func deleteTable(w http.ResponseWriter, r *http.Request) {
	tableName, client := extractTableContextValues(r)

	params := &dynamodb.DeleteTableInput{TableName: &tableName}
	_, err := client.DeleteTable(r.Context(), params)
	if err != nil {
		writeJSONResponse(w, statusCodeFromAWSError(err), NewErrResponse(err))
		return
	}

	writeJSONResponse(w, http.StatusOK, TableDeleteResponse{TableName: tableName})
}

func extractTableContextValues(r *http.Request) (string, *dynamodb.Client) {
	tableName := r.Context().Value(tableNameKey).(string)
	client := r.Context().Value(ddbClientKey).(*dynamodb.Client)
	return tableName, client
}

func statusCodeFromAWSError(err error) int {
	if strings.Contains(err.Error(), "ResourceNotFoundException") {
		return http.StatusNotFound
	}

	if strings.Contains(err.Error(), "AccessDeniedException") {
		return http.StatusForbidden
	}

	return http.StatusFailedDependency
}

func tableStructureParams(tableName string) dynamodb.CreateTableInput {
	return dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: &tableKeyPrimary,
				AttributeType: types.ScalarAttributeTypeN,
			},
			{
				AttributeName: &tableKeySorting,
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: &tableKeyPrimary,
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: &tableKeySorting,
				KeyType:       types.KeyTypeRange,
			},
		},
		TableName: &tableName,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		TableClass: types.TableClassStandardInfrequentAccess,
	}
}
