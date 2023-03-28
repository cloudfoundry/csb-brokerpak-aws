package acceptance_tests_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"csbbrokerpakaws/acceptance-tests/helpers/services"
)

type valueResponseType struct {
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
		createTablePayload := fmt.Sprintf(`{"table_name": "%s"}`, tableName)
		response := appOne.POST(createTablePayload, "/tables")
		Expect(response).To(HaveHTTPStatus(http.StatusAccepted))
		defer func() {
			_, _ = appOne.DELETEResponse("/tables/%s", tableName)
		}()

		By("storing a value in the created table")
		valuePayload := random.Name(random.WithPrefix("dynamodb-namespace-value"))
		valueSortKey := random.Name(random.WithPrefix("sort-key"))
		postBody, getBody := &valueResponseType{}, &valueResponseType{}

		// Table is reported as existing as soon as it is created,
		//  however, trying to create a value in it immediately results in a 404 error
		Eventually(func(g Gomega) {
			response, err := appOne.POSTResponse(valuePayload, "/tables/%s/values/%s", tableName, valueSortKey)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(response).To(HaveHTTPStatus(http.StatusCreated))
			bodyAsBytes, err := io.ReadAll(response.Body)
			g.Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(bodyAsBytes, postBody)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(postBody.Value).To(Equal(valuePayload))
			g.Expect(postBody.Sorting).To(Equal(valueSortKey))
			g.Expect(postBody.PK).To(BeNumerically(">", 0))
		}, "5m").Should(Succeed())

		By("checking the table presence using the second app")
		appTwo.GET("/tables/%s", tableName)

		By("reading the value using the second app")
		valueResponse := appTwo.GET("/tables/%s/values/%s/%d", tableName, valueSortKey, postBody.PK)
		err := json.Unmarshal([]byte(valueResponse), getBody)
		Expect(err).NotTo(HaveOccurred())
		Expect(*getBody).To(Equal(*postBody))

		By("destroying the table using the second app")
		appTwo.DELETE("/tables/%s", tableName)

		By("ensuring the table is gone eventually")
		Eventually(func(g Gomega) {
			getResponse, err := appTwo.GETResponse("/tables/%s", tableName)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(getResponse).To(HaveHTTPStatus(http.StatusNotFound))
		}, "5m").Should(Succeed())
	})
})
