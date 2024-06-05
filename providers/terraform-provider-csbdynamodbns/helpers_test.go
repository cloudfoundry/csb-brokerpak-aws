package main_test

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-dynamodbns/csbdynamodbns"
)

const testRegion = "us-west-2"

func populateTables(prefix string, client *dynamodb.Client) {
	tableNames := []string{
		"one",
		"two",
		"three",
		fmt.Sprintf("%sone", prefix),
		fmt.Sprintf("%stwo", prefix),
		fmt.Sprintf("%sthree", prefix),
	}
	for _, name := range tableNames {
		input := &dynamodb.CreateTableInput{
			TableName: &name,
			AttributeDefinitions: []types.AttributeDefinition{
				{AttributeName: ptr.String("pk"), AttributeType: types.ScalarAttributeTypeS},
			},
			KeySchema: []types.KeySchemaElement{
				{AttributeName: ptr.String("pk"), KeyType: types.KeyTypeHash},
			},
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  ptr.Int64(10),
				WriteCapacityUnits: ptr.Int64(10),
			},
		}
		_, err := client.CreateTable(context.TODO(), input)
		Expect(err).NotTo(HaveOccurred())
	}

	Eventually(func(g Gomega) {
		for _, name := range tableNames {
			input := &dynamodb.DescribeTableInput{TableName: &name}
			table, err := client.DescribeTable(context.TODO(), input)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(table.Table.TableStatus).To(Equal(types.TableStatusActive))
		}
	}).WithTimeout(time.Minute).WithPolling(5 * time.Second).Should(Succeed())
}

func applyHCL(hcl string, checkOnDestroy resource.TestCheckFunc) {
	resource.Test(GinkgoT(), resource.TestCase{
		IsUnitTest: true, // means we don't need to set TF_ACC
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"csbdynamodbns": func() (*schema.Provider, error) { return csbdynamodbns.Provider(), nil },
		},
		CheckDestroy: checkOnDestroy,
		Steps:        []resource.TestStep{{Config: hcl}},
	})
}
