package csbdynamodbns_test

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-dynamodbns/csbdynamodbns"
	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-dynamodbns/csbdynamodbns/csbdynamodbnsfakes"
)

var _ = Describe("ResourceDynamoDBNSInstance", func() {
	Expect(true)
	var (
		client *csbdynamodbnsfakes.FakeDynamoDBClient
		config *csbdynamodbnsfakes.FakeDynamoDBConfig
		data   *schema.ResourceData
	)

	BeforeEach(func() {
		client = &csbdynamodbnsfakes.FakeDynamoDBClient{}

		config = &csbdynamodbnsfakes.FakeDynamoDBConfig{}
		config.GetClientReturns(client, nil)
		config.GetPrefixReturns(fmt.Sprintf("csb-%s-", uuid.New()))

		data = csbdynamodbns.ResourceDynamoDBNSInstance().TestResourceData()
		Expect(data.Set(csbdynamodbns.AwsAccessKeyIDKey, "id")).NotTo(HaveOccurred())
		Expect(data.Set(csbdynamodbns.AwsSecretAccessKeyKey, "key")).NotTo(HaveOccurred())

	})

	Context("various tables exist", func() {
		BeforeEach(func() {
			prefix := config.GetPrefix()
			secondUUID, thirdUUID := uuid.New(), uuid.New()
			client.ListTablesReturns(&dynamodb.ListTablesOutput{TableNames: []string{
				fmt.Sprintf("%s-one", prefix),
				fmt.Sprintf("csb-%s-two", secondUUID),
				fmt.Sprintf("csb-%s-three", thirdUUID),
				fmt.Sprintf("%s-four", prefix),
				fmt.Sprintf("csb-%s-five", secondUUID),
				fmt.Sprintf("csb-%s-six", thirdUUID),
				fmt.Sprintf("%s-seven", prefix),
				fmt.Sprintf("csb-%s-eight", secondUUID),
				fmt.Sprintf("csb-%s-nine", thirdUUID),
			}}, nil)
		})

		It("runs delete for every returned table with the given prefix", func() {
			d := csbdynamodbns.ResourceDynamoDBMaintenanceDelete(context.TODO(), data, config)
			Expect(d).To(BeNil())
			Expect(client.ListTablesCallCount()).To(Equal(1))

			Expect(client.DeleteTableCallCount()).To(Equal(3))
		})

		It("accumulates table deletion errors", func() {
			client.DeleteTableReturnsOnCall(0, nil, fmt.Errorf("table 0 deletion failed"))
			client.DeleteTableReturnsOnCall(2, nil, fmt.Errorf("table 2 deletion failed"))

			d := csbdynamodbns.ResourceDynamoDBMaintenanceDelete(context.TODO(), data, config)
			Expect(d).NotTo(BeNil())
			Expect(d.HasError()).To(BeTrue())
			Expect(d).To(HaveLen(2))
			Expect(d[0].Summary).To(Equal("table 0 deletion failed"))
			Expect(d[1].Summary).To(Equal("table 2 deletion failed"))
		})
	})

	Context("multi-page output with various errors", func() {
		BeforeEach(func() {
			prefix := config.GetPrefix()
			client.ListTablesReturnsOnCall(0, &dynamodb.ListTablesOutput{
				TableNames: []string{
					fmt.Sprintf("%s-etiam-ducunt-ad-castus-nuptia", prefix),
					"rum-eggs-bacon-and-spam",
					fmt.Sprintf("%s-politics-ye-fine-anchor!", prefix),
					fmt.Sprintf("%s-creature-tribbles-question-fantastic", prefix),
					"mossiness-of-mackerel-mousse",
				},
				LastEvaluatedTableName: ptr.String("mossiness-of-mackerel-mousse"),
			}, nil)
			client.ListTablesReturnsOnCall(1, nil, fmt.Errorf("connection issues"))
			client.DeleteTableReturnsOnCall(0, nil, fmt.Errorf("table 0 deletion failed"))
		})

		It("reports all the errors", func() {
			d := csbdynamodbns.ResourceDynamoDBMaintenanceDelete(context.TODO(), data, config)
			Expect(d).NotTo(BeNil())
			Expect(d).To(HaveLen(2))
			Expect(d[0].Summary).To(Equal("table 0 deletion failed"))
			Expect(d[1].Summary).To(Equal("connection issues"))

			Expect(client.ListTablesCallCount()).To(Equal(2))
		})
	})

	It("Reports the client errors", func() {
		client.ListTablesReturns(nil, fmt.Errorf("ouch"))

		d := csbdynamodbns.ResourceDynamoDBMaintenanceDelete(context.TODO(), data, config)
		Expect(d).NotTo(BeNil())
		Expect(d.HasError()).To(BeTrue())
		Expect(d[0].Summary).To(Equal("ouch"))
	})

})
