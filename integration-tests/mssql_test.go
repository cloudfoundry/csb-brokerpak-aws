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
	var requiredProperties = func() map[string]any {
		return map[string]any{
			"engine":        "sqlserver-ee",
			"mssql_version": "some-mssql-version",
			// For documentation purpose:
			// use a valid storage GB value. Default iops is 3000.
			// The IOPS to GiB ratio must be between 1 and 50
			"storage_gb": 100,

			"instance_class": "some-instance-class",
		}
	}

	var optionalProperties = func() map[string]any {
		return map[string]any{
			"rds_vpc_security_group_ids": "some-security-group-ids",
			"rds_subnet_group":           "some-rds-subnet-group",
			"instance_class":             "some-instance-class",
			"max_allocated_storage":      999,
		}
	}

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
		Expect(service.Metadata.DocumentationUrl).To(Equal(documentationURL))
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
		DescribeTable("required properties",
			func(property string) {
				_, err := broker.Provision(msSQLServiceName, "custom-sample", deleteProperty(property, requiredProperties()))

				Expect(err).To(MatchError(ContainSubstring("1 error(s) occurred: (root): " + property + " is required")))
			},
			Entry("engine is required", "engine"),
			Entry("mssql_version is required", "mssql_version"),
			Entry("storage_gb is required", "storage_gb"),
		)

		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(msSQLServiceName, "custom-sample", buildProperties(requiredProperties(), params))

				Expect(err).To(MatchError(ContainSubstring(`1 error(s) occurred: ` + expectedErrorMsg)))
			},
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html#options
				"instance name will be used as db-instance-identifier so must contain from 1 to 63 letters, numbers or hyphens",
				map[string]any{"instance_name": stringOfLen(64)},
				"instance_name: String length must be less than or equal to 63",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html#options
				"instance name will be used as db-instance-identifier so the first character must be a letter",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-zA-Z](-?[a-zA-Z0-9])*$'",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html#options
				"instance name will be used as db-instance-identifier so it cannot end with a hyphen",
				map[string]any{"instance_name": "aaaaa-"},
				"instance_name: Does not match pattern '^[a-zA-Z](-?[a-zA-Z0-9])*$'",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html#options
				"instance name will be used as db-instance-identifier so it cannot contain two consecutive hyphens",
				map[string]any{"instance_name": "aa--aaa"},
				"instance_name: Does not match pattern '^[a-zA-Z](-?[a-zA-Z0-9])*$'",
			),
			Entry(
				"database name maximum length is 64 characters",
				map[string]any{"db_name": stringOfLen(65)},
				"db_name: String length must be less than or equal to 64",
			),
			Entry(
				"engine must be one of the allowed values",
				map[string]any{"engine": "not-an-allowed-engine"},
				`engine: engine must be one of the following: \"sqlserver-ee\", \"sqlserver-ex\", \"sqlserver-se\", \"sqlserver-web\"`,
			),
			Entry(
				"storage_gb minimum value is 20",
				map[string]any{"storage_gb": 19},
				"storage_gb: Must be greater than or equal to 20",
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(msSQLServiceName, customMSSQLPlan["name"].(string), buildProperties(requiredProperties(), optionalProperties()))
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{
						"pcf-instance-id": Equal(instanceID),
						"key1":            Equal("value1"),
						"key2":            Equal("value2"),
					})),
					HaveKeyWithValue("engine", "sqlserver-ee"),
					HaveKeyWithValue("mssql_version", "some-mssql-version"),
					HaveKeyWithValue("storage_gb", BeNumerically("==", 100)),
					HaveKeyWithValue("rds_subnet_group", "some-rds-subnet-group"),
					HaveKeyWithValue("rds_vpc_security_group_ids", "some-security-group-ids"),
					HaveKeyWithValue("instance_class", "some-instance-class"),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-mssql-%s", instanceID)),
					HaveKeyWithValue("storage_encrypted", BeTrue()),
					HaveKeyWithValue("kms_key_id", ""),
					HaveKeyWithValue("db_name", "vsbdb"),
					HaveKeyWithValue("region", fakeRegion),
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{"pcf-instance-id": Equal(instanceID)})),
					HaveKeyWithValue("max_allocated_storage", BeNumerically("==", 999)),
					HaveKeyWithValue("storage_type", "io1"),
					HaveKeyWithValue("iops", BeNumerically("==", 3000)),
					HaveKeyWithValue("deletion_protection", BeFalse()),
					HaveKeyWithValue("publicly_accessible", BeFalse()),
					HaveKeyWithValue("monitoring_interval", BeNumerically("==", 0)),
					HaveKeyWithValue("monitoring_role_arn", Equal("")),
				),
			)
		})

	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(msSQLServiceName, customMSSQLPlan["name"].(string), requiredProperties())

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
			Entry("update storage_encrypted", "storage_encrypted", false),
			Entry("update kms_key_id", "kms_key_id", "no-matter-what-kms-key-id"),
			Entry("update db_name", "db_name", "no-matter-what-name"),
			Entry("update instance_name", "instance_name", "no-matter-what-instance-name"),
			Entry("update rds_vpc_security_group_ids", "rds_vpc_security_group_ids", "no-matter-what-security-group"),
		)

		DescribeTable("should allow unsetting properties flagged as `nullable` by explicitly updating their value to be `nil`",
			func(prop string, initValue any) {
				err := broker.Update(instanceID, msSQLServiceName, customMSSQLPlan["name"].(string), map[string]any{prop: initValue})
				Expect(err).NotTo(HaveOccurred())

				err = broker.Update(instanceID, msSQLServiceName, customMSSQLPlan["name"].(string), map[string]any{prop: nil})
				Expect(err).NotTo(HaveOccurred())
			},
			Entry("max_allocated_storage is nullable", "max_allocated_storage", 999),
		)

		DescribeTable("should prevent unsetting properties not flagged as `nullable` by explicitly updating their value to be `nil`",
			func(prop string, initValue any) {
				err := broker.Update(instanceID, msSQLServiceName, customMSSQLPlan["name"].(string), map[string]any{prop: initValue})
				Expect(err).NotTo(HaveOccurred())

				err = broker.Update(instanceID, msSQLServiceName, customMSSQLPlan["name"].(string), map[string]any{prop: nil})
				Expect(err).To(MatchError(ContainSubstring("Invalid type. Expected: ")))
				Expect(err).To(MatchError(ContainSubstring(", given: null")))
			},
			Entry("rds_subnet_group isn't nullable", "rds_subnet_group", "any-value"),
			Entry("publicly_accessible isn't nullable", "publicly_accessible", true),
		)

		DescribeTable("should allow updating properties",
			func(prop string, value any) {
				err := broker.Update(instanceID, msSQLServiceName, customMySQLPlan["name"].(string), map[string]any{prop: value})

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update storage_type", "storage_type", "gp2"),
			Entry("update iops", "iops", 1500),
			Entry("update deletion_protection", "deletion_protection", true),
			Entry("update publicly_accessible", "publicly_accessible", true),
		)
	})
})
