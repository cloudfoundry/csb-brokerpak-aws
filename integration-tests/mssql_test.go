package integration_test

import (
	"fmt"

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

	Describe("provisioning", func() {
		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(msSQLServiceName, "custom-sample", params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"instance name minimum length is 6 characters",
				map[string]any{"instance_name": stringOfLen(5)},
				"instance_name: String length must be greater than or equal to 6",
			),
			Entry(
				"instance name maximum length is 98 characters",
				map[string]any{"instance_name": stringOfLen(99)},
				"instance_name: String length must be less than or equal to 98",
			),
			Entry(
				"instance name invalid characters",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"database name maximum length is 64 characters",
				map[string]any{"db_name": stringOfLen(65)},
				"db_name: String length must be less than or equal to 64",
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(msSQLServiceName, customMSSQLPlan["name"].(string), nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
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
			instanceID, err = broker.Provision(msSQLServiceName, customMSSQLPlan["name"].(string), nil)

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance",
			func(prop string, value any) {
				err := broker.Update(instanceID, msSQLServiceName, customMSSQLPlan["name"].(string), map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update region", "region", "no-matter-what-region"),
			Entry("update db_name", "db_name", "no-matter-what-name"),
		)
	})
})
