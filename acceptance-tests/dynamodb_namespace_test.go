package acceptance_tests_test

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

type dynamoDBValueResponseType struct {
	PK      int    `json:"pk"`
	Sorting string `json:"sorting"`
	Value   string `json:"value"`
}

var _ = Describe("DynamoDB Namespace", Label("dynamodb-namespace"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance(
			"csb-aws-dynamodb-namespace",
			services.WithPlan("default"),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.DynamoDBNamespace))
		appTwo := apps.Push(apps.WithApp(apps.DynamoDBNamespace))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the DynamoDB Table service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(HaveKey("credhub-ref"))

		By("creating a table using the prefix")
		tableName := fmt.Sprintf("csb-%s-%s", serviceInstance.GUID(), random.Hexadecimal())
		createTablePayload := map[string]string{"table_name": tableName}
		appOne.POST(createTablePayload, "/tables")

		By("storing a value in the created table")
		valuePayload := random.Name(random.WithPrefix("dynamodb-namespace-value"))
		valueSortKey := random.Name(random.WithPrefix("sort-key"))

		// Table is reported as existing as soon as it is created,
		// however, trying to create a value in it immediately results in a 404 error
		var postBody dynamoDBValueResponseType
		Eventually(func(g Gomega) {
			postResponse := appOne.POSTResponse(valuePayload, "/tables/%s/values/%s", tableName, valueSortKey)
			g.Expect(postResponse).To(HaveHTTPStatus(http.StatusCreated))
			defer postResponse.Body.Close()
			apps.NewPayload(postResponse).ParseInto(&postBody)
		}).WithTimeout(5 * time.Minute).WithPolling(time.Second).Should(Succeed())

		Expect(postBody.Value).To(Equal(valuePayload))
		Expect(postBody.Sorting).To(Equal(valueSortKey))
		Expect(postBody.PK).To(BeNumerically(">", 0))

		By("checking the table presence using the second app")
		appTwo.GET("/tables/%s", tableName)

		By("reading the value using the second app")
		var getBody dynamoDBValueResponseType
		appTwo.GET("/tables/%s/values/%s/%d", tableName, valueSortKey, postBody.PK).ParseInto(&getBody)
		Expect(getBody).To(Equal(postBody))

		By("destroying the table using the second app")
		appTwo.DELETE("/tables/%s", tableName)

		By("ensuring the table is gone eventually")
		Eventually(func(g Gomega) {
			getResponse := appTwo.GETResponse("/tables/%s", tableName)
			g.Expect(getResponse).To(HaveHTTPStatus(http.StatusNotFound))
		}).WithTimeout(5 * time.Minute).WithPolling(time.Second).Should(Succeed())
	})
})
