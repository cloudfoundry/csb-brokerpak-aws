package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/awscli"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

var _ = Describe("DynamoDB Namespace Data Migration", Label("dynamodb-namespace-migration"), func() {
	It("can migrate data from the legacy broker", func() {
		By("creating a legacy service instance")
		legacyServiceInstance := services.CreateInstance(
			"aws-dynamodb",
			services.WithPlan("standard"),
			services.WithBrokerName("aws-services-broker"),
		)
		defer legacyServiceInstance.Delete()

		By("creating a service CSB service instance")
		csbServiceInstance := services.CreateInstance(
			"csb-aws-dynamodb-namespace",
			services.WithPlan("default"),
		)
		defer csbServiceInstance.Delete()

		By("pushing and binding an app")
		app := apps.Push(apps.WithApp(apps.DynamoDBNamespace))
		defer apps.Delete(app)
		legacyBinding := legacyServiceInstance.Bind(app)
		apps.Start(app)

		By("creating a table")
		legacyPrefix := fmt.Sprintf("%s_", legacyServiceInstance.GUID())
		tableSuffix := random.Hexadecimal()
		legacyTable := fmt.Sprintf("%s%s", legacyPrefix, tableSuffix)
		app.POST(map[string]string{"table_name": legacyTable}, "/tables")

		By("storing a value in the created table")
		valuePayload := random.Name(random.WithPrefix("dynamodb-namespace-value"))
		valueSortKey := random.Name(random.WithPrefix("sort-key"))

		// Table is reported as existing as soon as it is created,
		// however, trying to create a value in it immediately results in a 404 error
		var postBody dynamoDBValueResponseType
		Eventually(func(g Gomega) {
			postResponse := app.POSTResponse(valuePayload, "/tables/%s/values/%s", legacyTable, valueSortKey)
			g.Expect(postResponse).To(HaveHTTPStatus(http.StatusCreated))
			defer postResponse.Body.Close()
			apps.NewPayload(postResponse).ParseInto(&postBody)
		}).WithTimeout(5 * time.Minute).WithPolling(time.Second).Should(Succeed())

		Expect(postBody.Value).To(Equal(valuePayload))
		Expect(postBody.Sorting).To(Equal(valueSortKey))
		Expect(postBody.PK).To(BeNumerically(">", 0))

		By("creating a backup of the table")
		backupName := random.Name(random.WithPrefix("dynamodb-backup"))
		var backupReceiver struct {
			ARN string `jsonry:"BackupDetails.BackupArn"`
		}
		awscli.AWSToJSON(
			&backupReceiver,
			"dynamodb", "create-backup",
			"--table-name", legacyTable,
			"--backup-name", backupName,
			"--region", metadata.Region,
		)

		Eventually(func(g Gomega) {
			var describeReceiver struct {
				Status string `jsonry:"BackupDescription.BackupDetails.BackupStatus"`
			}
			awscli.AWSToJSON(&describeReceiver, "dynamodb", "describe-backup", "--backup-arn", backupReceiver.ARN, "--region", metadata.Region)
			g.Expect(describeReceiver.Status).To(Equal("AVAILABLE"))
		}).WithTimeout(time.Hour).WithPolling(time.Second).Should(Succeed())

		By("destroying the table")
		app.DELETE("/tables/%s", legacyTable)

		By("restoring the table into the namespace of the CSB service")
		csbTableName := fmt.Sprintf("csb-%s-%s", csbServiceInstance.GUID(), tableSuffix)
		awscli.AWS("dynamodb", "restore-table-from-backup", "--target-table-name", csbTableName, "--backup-arn", backupReceiver.ARN, "--region", metadata.Region)

		Eventually(func(g Gomega) {
			var describeReceiver struct {
				RestoreInProgress bool   `jsonry:"Table.RestoreSummary.RestoreInProgress"`
				Status            string `jsonry:"Table.TableStatus"`
			}
			awscli.AWSToJSON(&describeReceiver, "dynamodb", "describe-table", "--table-name", csbTableName, "--region", metadata.Region)
			g.Expect(describeReceiver.Status).To(Equal("ACTIVE"))
			g.Expect(describeReceiver.RestoreInProgress).To(BeFalse())
		}).WithTimeout(time.Hour).WithPolling(10 * time.Second).Should(Succeed())

		By("unbinding and rebinding the app")
		legacyBinding.Unbind()
		csbServiceInstance.Bind(app)
		app.Restage()

		By("checking the table presence using the app")
		app.GET("/tables/%s", csbTableName)

		By("reading the value using the app")
		var getBody dynamoDBValueResponseType
		app.GET("/tables/%s/values/%s/%d", csbTableName, valueSortKey, postBody.PK).ParseInto(&getBody)
		Expect(getBody).To(Equal(postBody))

		By("destroying the table using the app")
		app.DELETE("/tables/%s", csbTableName)

		By("ensuring the table is gone eventually")
		Eventually(func(g Gomega) {
			getResponse := app.GETResponse("/tables/%s", csbTableName)
			g.Expect(getResponse).To(HaveHTTPStatus(http.StatusNotFound))
		}).WithTimeout(5 * time.Minute).WithPolling(time.Second).Should(Succeed())
	})
})
