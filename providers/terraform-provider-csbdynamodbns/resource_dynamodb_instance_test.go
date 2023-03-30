package main_test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/ptr"
	"github.com/cloudfoundry/cloud-service-broker/utils/freeport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-dynamodbns/csbdynamodbns"
)

var _ = Describe("Resource dynamodbns_maintenance", func() {
	var session *gexec.Session
	var port int
	var localDynamoDBURL string
	var client *dynamodb.Client
	var prefix string

	BeforeEach(func() {
		var err error

		port = freeport.Must()
		localDynamoDBURL = fmt.Sprintf("http://localhost:%d", port)
		prefix = "csb-0d493f5d-24b4-4e49-8417-90abd3c0c1c0"

		cmd := exec.Command("docker", "run",
			"-p", fmt.Sprintf("%d:8000", port),
			"-t", "amazon/dynamodb-local:1.21.0")

		GinkgoWriter.Printf("running command: %s\n", cmd)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) {
			cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(
					func(service, region string, options ...interface{}) (aws.Endpoint, error) {
						return aws.Endpoint{URL: localDynamoDBURL}, nil
					},
				),
			))
			g.Expect(err).NotTo(HaveOccurred())
			client = dynamodb.NewFromConfig(cfg)
			_, err = client.ListTables(context.TODO(), nil)
			g.Expect(err).NotTo(HaveOccurred())

		}).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(Succeed())

		tableNames := []string{
			"one",
			"two",
			"three",
			fmt.Sprintf("%s-one", prefix),
			fmt.Sprintf("%s-two", prefix),
			fmt.Sprintf("%s-three", prefix),
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

	})

	AfterEach(func() {
		session.Terminate()
	})

	It("should clean up", func() {
		tfBody := fmt.Sprintf(`provider "csbdynamodbns" {
		 region              = "us-west-2"
		 prefix              = "%s"
		 custom_endpoint_url = "%s"
		}
		
		resource "csbdynamodbns_instance" "service_instance" {
		 access_key_id     = "fake-key-id"
		 secret_access_key = "fake-secret-key"
		}
		`, prefix, localDynamoDBURL)
		applyHCL(tfBody, func(state *terraform.State) error {
			By("checking that only non-prefixed tables remain")

			tables, err := client.ListTables(context.TODO(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(tables.TableNames).To(HaveLen(3))
			return nil
		})
	})
})

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
