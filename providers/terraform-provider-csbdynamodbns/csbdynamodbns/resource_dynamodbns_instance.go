package csbdynamodbns

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	AwsAccessKeyIDKey     = "access_key_id"
	AwsSecretAccessKeyKey = "secret_access_key"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -header csbdynamodbnsfakes/header.txt . DynamoDBClient
type DynamoDBClient interface {
	dynamodb.ListTablesAPIClient
	DeleteTable(context.Context, *dynamodb.DeleteTableInput, ...func(options *dynamodb.Options)) (*dynamodb.DeleteTableOutput, error)
}

var _ DynamoDBClient = &dynamodb.Client{}

func ResourceDynamoDBNSInstance() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			AwsAccessKeyIDKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			AwsSecretAccessKeyKey: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		CreateContext: setResourceID,
		UpdateContext: setResourceID,
		ReadContext:   setResourceID,
		DeleteContext: ResourceDynamoDBMaintenanceDelete,
		Description:   "Handles DynamoDB namespace housekeeping",
	}
}

func setResourceID(_ context.Context, data *schema.ResourceData, _ any) diag.Diagnostics {
	keyID := data.Get(AwsAccessKeyIDKey).(string)
	data.SetId(keyID)
	return nil
}

func ResourceDynamoDBMaintenanceDelete(ctx context.Context, data *schema.ResourceData, config any) diag.Diagnostics {
	settings := config.(DynamoDBConfig)
	client, err := settings.GetClient(ctx, data.Get(AwsAccessKeyIDKey).(string), data.Get(AwsSecretAccessKeyKey).(string))
	if err != nil {
		return diag.FromErr(err)
	}
	paginator := dynamodb.NewListTablesPaginator(client, &dynamodb.ListTablesInput{}, func(o *dynamodb.ListTablesPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	d := diag.Diagnostics{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			// We have to return immediately in order to avoid an infinite loop
			return append(d, diag.Diagnostic{Severity: diag.Error, Summary: err.Error()})
		}
		for _, tableName := range page.TableNames {
			if strings.HasPrefix(tableName, settings.GetPrefix()) {
				_, err := client.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: ptr.String(tableName)})
				if err != nil {
					d = append(d, diag.Diagnostic{Severity: diag.Error, Summary: err.Error()})
				}
			}
		}
	}
	if len(d) > 0 {
		return d
	}
	return nil
}
