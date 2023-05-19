package integration_test

import (
	"fmt"

	"golang.org/x/exp/maps"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	msSQLServiceID                  = "8b17758e-37a9-4c1c-af84-971d4a5552c1"
	msSQLServiceName                = "csb-aws-mssql"
	msSQLServiceDescription         = "CSB Amazon RDS for MSSQL"
	msSQLServiceDisplayName         = "CSB Amazon RDS for MSSQL"
	msSQLServiceDocumentationURL    = "https://docs.vmware.com/en/Cloud-Service-Broker-for-VMware-Tanzu/index.html"
	msSQLServiceSupportURL          = "https://aws.amazon.com/sql/"
	msSQLServiceProviderDisplayName = "VMware"
	msSQLCustomPlanName             = "custom-sample"
	msSQLCustomPlanID               = "d10e9572-0ea8-4bad-a4f3-a9a084dde067"
)

var customMSSQLPlans = []map[string]any{
	customMSSQLPlan,
}

var customMSSQLPlan = map[string]any{
	"name":        msSQLCustomPlanName,
	"id":          msSQLCustomPlanID,
	"description": "Default MSSQL plan",
	"metadata": map[string]any{
		"displayName": "custom-sample",
	},
}

var requiredProperties = map[string]any{
	"mssql_version": "some-mssql-version",
	"storage_gb":    123,
}

var defaultProperties = map[string]any{
	"engine":        "sqlserver-ee",
	"region":        "us-west-2",
	"instance_name": "csb-mssql-0000000",
	"db_name":       "vsbdb",
}

var optionalProperties = map[string]any{
	"rds_vpc_security_group_ids": "some-security-group-ids",
	"rds_subnet_group":           "some-rds-subnet-group",
	"instance_class":             "some-instance-class",
}

var _ = Describe("MSSQL", Label("MSSQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish AWS MSSQL in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, msSQLServiceName)
		Expect(service.ID).To(Equal(msSQLServiceID))
		Expect(service.Description).To(Equal(msSQLServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "mssql", "beta"))
		Expect(service.Metadata.DisplayName).To(Equal(msSQLServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(msSQLServiceDocumentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(msSQLServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(msSQLServiceProviderDisplayName))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal(msSQLCustomPlanName),
					ID:   Equal(msSQLCustomPlanID),
				}),
			),
		)
	})

	Describe("provisioning without specifying required properties", func() {
		It("should fail", func() {
			_, err := broker.Provision(msSQLServiceName, customMSSQLPlan["name"].(string), nil)
			Expect(err.Error()).To(Equal(
				`unexpected status code 500: {"description":"2 error(s) occurred: (root): mssql_version is required; (root): storage_gb is required"}` + "\n",
			))
		})
	})

	Describe("provisioning", func() {
		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				// Combine required properties and received params
				// In case of collision params take precedence
				combinedProperties := map[string]any{}
				maps.Copy(combinedProperties, requiredProperties)
				maps.Copy(combinedProperties, defaultProperties)
				maps.Copy(combinedProperties, params)
				// If we included required properties directly in the plan we wouldn't be able to test
				// certain scenarios easily. For example, passing an invalid engine value would require
				// another test and another plan without the engine (user-inputs can't override plan properties)
				//
				// However, with this "COMPOSITIONAL" approach we have more fine-grained control over the inputs

				_, err := broker.Provision(msSQLServiceName, "custom-sample", combinedProperties)

				Expect(err).To(MatchError(expectedErrorMsg))
			},
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				`unexpected status code 500: {"description":"1 error(s) occurred: region: Does not match pattern '^[a-z][a-z0-9-]+$'"}`+"\n",
			),
			Entry(
				"instance name minimum length is 6 characters",
				map[string]any{"instance_name": stringOfLen(5)},
				`unexpected status code 500: {"description":"1 error(s) occurred: instance_name: String length must be greater than or equal to 6"}`+"\n",
			),
			Entry(
				"instance name maximum length is 98 characters",
				map[string]any{"instance_name": stringOfLen(99)},
				`unexpected status code 500: {"description":"1 error(s) occurred: instance_name: String length must be less than or equal to 98"}`+"\n",
			),
			Entry(
				"instance name invalid characters",
				map[string]any{"instance_name": ".aaaaa"},
				`unexpected status code 500: {"description":"1 error(s) occurred: instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'"}`+"\n",
			),
			Entry(
				"database name maximum length is 64 characters",
				map[string]any{"db_name": stringOfLen(65)},
				`unexpected status code 500: {"description":"1 error(s) occurred: db_name: String length must be less than or equal to 64"}`+"\n",
			),
			Entry(
				"engine must be one of the allowed values",
				map[string]any{"engine": "not-an-allowed-engine"},
				`unexpected status code 500: {"description":"1 error(s) occurred: engine: engine must be one of the following: \"sqlserver-ee\", \"sqlserver-ex\", \"sqlserver-se\", \"sqlserver-web\""}`+"\n",
			),
		)

		It("should provision a plan", func() {
			combinedProperties := map[string]any{}
			maps.Copy(combinedProperties, requiredProperties)
			maps.Copy(combinedProperties, optionalProperties)
			instanceID, err := broker.Provision(msSQLServiceName, customMSSQLPlan["name"].(string), combinedProperties)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("engine", "sqlserver-ee"),
					HaveKeyWithValue("mssql_version", "some-mssql-version"),
					HaveKeyWithValue("storage_gb", float64(123)),
					HaveKeyWithValue("rds_subnet_group", "some-rds-subnet-group"),
					HaveKeyWithValue("rds_vpc_security_group_ids", "some-security-group-ids"),
					HaveKeyWithValue("instance_class", "some-instance-class"),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-mssql-%s", instanceID)),
					HaveKeyWithValue("db_name", "vsbdb"),
					HaveKeyWithValue("region", fakeRegion),
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{"pcf-instance-id": Equal(instanceID)})),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(msSQLServiceName, customMSSQLPlan["name"].(string), requiredProperties)

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance",
			func(prop string, value any) {
				err := broker.Update(instanceID, msSQLServiceName, customMSSQLPlan["name"].(string), map[string]any{prop: value})

				Expect(err.Error()).To(Equal(
					`unexpected status code 400: {"description":"attempt to update parameter that may result in service instance re-creation and data loss"}` + "\n",
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update region", "region", "no-matter-what-region"),
			Entry("update db_name", "db_name", "no-matter-what-name"),
			Entry("update instance_name", "instance_name", "no-matter-what-instance-name"),
			Entry("update rds_vpc_security_group_ids", "rds_vpc_security_group_ids", "no-matter-what-security-group"),
		)
	})
})
