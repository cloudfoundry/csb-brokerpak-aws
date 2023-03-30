package csbdynamodbns

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	awsAccessKeyIDKey     = "access_key_id"
	awsSecretAccessKeyKey = "secret_access_key"
)

func resourceDynamoDBNSInstance() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			awsAccessKeyIDKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			awsSecretAccessKeyKey: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		CreateContext: setResourceID,
		UpdateContext: setResourceID,
		ReadContext:   setResourceID,
		DeleteContext: resourceDynamoDBMaintenanceDelete,
		Description:   "Handles DynamoDB namespace housekeeping",
	}
}

func setResourceID(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	keyID := d.Get(awsAccessKeyIDKey).(string)
	d.SetId(keyID)
	return nil
}

func resourceDynamoDBMaintenanceDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	settings := m.(*dynamoDBNamespaceSettings)
	client, err := settings.GetClient(ctx, d.Get(awsAccessKeyIDKey).(string), d.Get(awsSecretAccessKeyKey).(string))
	if err != nil {
		return diag.FromErr(err)
	}
	paginator := dynamodb.NewListTablesPaginator(client, &dynamodb.ListTablesInput{}, func(o *dynamodb.ListTablesPaginatorOptions) {
		// This should not even be a setting
		o.StopOnDuplicateToken = true
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		var multiErr *multierror.Error
		for _, tableName := range page.TableNames {
			if strings.HasPrefix(tableName, settings.prefix) {
				_, err := client.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: ptr.String(tableName)})
				if err != nil {
					multiErr = multierror.Append(multiErr, err)
				}
			}
		}
		if multiErr != nil {
			return diag.FromErr(multiErr)
		}
	}
	return nil
}
