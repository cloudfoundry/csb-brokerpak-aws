package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-chi/render"
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

	RenderableResponse
}

type TableDeleteResponse struct {
	TableName string `json:"table_name"`

	RenderableResponse
}

type TableGetResponse struct {
	TableName string `json:"table_name"`

	RenderableResponse
}

func getTable(w http.ResponseWriter, r *http.Request) {
	tableName, client := extractTableContextValues(r)

	reply, err := client.DescribeTable(r.Context(), &dynamodb.DescribeTableInput{TableName: &tableName})
	if err != nil {
		_ = render.Render(w, r, errorResponseFromAWSError(err))
		return
	}
	_ = render.Render(w, r, &TableGetResponse{TableName: *reply.Table.TableName})
}

func createTable(w http.ResponseWriter, r *http.Request) {
	client := r.Context().Value(ddbClientKey).(*dynamodb.Client)
	createTableRequestBytes, err := io.ReadAll(r.Body)
	if err != nil {
		_ = render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: err.Error()})
		return
	}
	var createTableRequest CreateTableRequest
	if err = json.Unmarshal(createTableRequestBytes, &createTableRequest); err != nil {
		_ = render.Render(w, r, &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: err.Error()})
		return
	}

	params := tableStructureParams(createTableRequest.TableName)
	table, err := client.CreateTable(r.Context(), &params)
	if err != nil {
		_ = render.Render(w, r, errorResponseFromAWSError(err))
		return
	}
	response := TableCreateResponse{
		TableName:  *table.TableDescription.TableName,
		CreatedAt:  table.TableDescription.CreationDateTime.String(),
		TableClass: string(table.TableDescription.TableClassSummary.TableClass),
	}
	render.Status(r, http.StatusAccepted)
	_ = render.Render(w, r, &response)
}

func deleteTable(w http.ResponseWriter, r *http.Request) {
	tableName, client := extractTableContextValues(r)

	params := &dynamodb.DeleteTableInput{TableName: &tableName}
	table, err := client.DeleteTable(r.Context(), params)
	if err != nil {
		_ = render.Render(w, r, errorResponseFromAWSError(err))
		return
	}
	response := TableDeleteResponse{TableName: *table.TableDescription.TableName}
	_ = render.Render(w, r, &response)
}

func extractTableContextValues(r *http.Request) (string, *dynamodb.Client) {
	tableName := r.Context().Value(tableNameKey).(string)
	client := r.Context().Value(ddbClientKey).(*dynamodb.Client)
	return tableName, client
}

func errorResponseFromAWSError(err error) *ErrResponse {
	statusCode := http.StatusFailedDependency
	if strings.Contains(err.Error(), "ResourceNotFoundException") {
		statusCode = http.StatusNotFound
	} else if strings.Contains(err.Error(), "AccessDeniedException") {
		statusCode = http.StatusForbidden
	}
	return &ErrResponse{HTTPStatusCode: statusCode, StatusText: err.Error()}
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
			ReadCapacityUnits:  int64Ptr(1),
			WriteCapacityUnits: int64Ptr(1),
		},
		TableClass: types.TableClassStandardInfrequentAccess,
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}
