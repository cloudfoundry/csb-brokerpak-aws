package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
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
			// use a valid storage GB value. Default iops is 1000.
			// The IOPS to GiB ratio must be between 1 and 50
			"storage_gb": 100,

			"instance_class": "some-instance-class",
		}
	}

	var optionalProperties = func() map[string]any {
		return map[string]any{
			"rds_vpc_security_group_ids":                   "some-security-group-ids",
			"rds_subnet_group":                             "some-rds-subnet-group",
			"instance_class":                               "some-instance-class",
			"max_allocated_storage":                        999,
			"auto_minor_version_upgrade":                   false,
			"allow_major_version_upgrade":                  false,
			"enable_export_agent_logs":                     true,
			"cloudwatch_agent_log_group_retention_in_days": 1,
			"enable_export_error_logs":                     true,
			"cloudwatch_error_log_group_retention_in_days": 1,
			"cloudwatch_log_groups_kms_key_id":             "arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx",
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
		Expect(service.Tags).To(ConsistOf("aws", "mssql"))
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
				"instance_name: Does not match pattern '^[a-z](-?[a-z0-9])*$'",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html#options
				"instance name will be used as db-instance-identifier so it cannot end with a hyphen",
				map[string]any{"instance_name": "aaaaa-"},
				"instance_name: Does not match pattern '^[a-z](-?[a-z0-9])*$'",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-instance.html#options
				"instance name will be used as db-instance-identifier so it cannot contain two consecutive hyphens",
				map[string]any{"instance_name": "aa--aaa"},
				"instance_name: Does not match pattern '^[a-z](-?[a-z0-9])*$'",
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
			Entry(
				"cloudwatch_agent_log_group_retention_in_days minimum value is 0",
				map[string]any{"cloudwatch_agent_log_group_retention_in_days": -1},
				"cloudwatch_agent_log_group_retention_in_days: Must be greater than or equal to 0",
			),
			Entry(
				"cloudwatch_agent_log_group_retention_in_days maximum value is 3653",
				map[string]any{"cloudwatch_agent_log_group_retention_in_days": 3654},
				"cloudwatch_agent_log_group_retention_in_days: Must be less than or equal to 3653",
			),
			Entry(
				"cloudwatch_error_log_group_retention_in_days minimum value is 0",
				map[string]any{"cloudwatch_error_log_group_retention_in_days": -1},
				"cloudwatch_error_log_group_retention_in_days: Must be greater than or equal to 0",
			),
			Entry(
				"cloudwatch_error_log_group_retention_in_days maximum value is 3653",
				map[string]any{"cloudwatch_error_log_group_retention_in_days": 3654},
				"cloudwatch_error_log_group_retention_in_days: Must be less than or equal to 3653",
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
					HaveKeyWithValue("iops", BeNumerically("==", 1000)),
					HaveKeyWithValue("deletion_protection", BeFalse()),
					HaveKeyWithValue("publicly_accessible", BeFalse()),
					HaveKeyWithValue("monitoring_interval", BeNumerically("==", 0)),
					HaveKeyWithValue("monitoring_role_arn", Equal("")),
					HaveKeyWithValue("backup_retention_period", float64(7)),
					HaveKeyWithValue("copy_tags_to_snapshot", true),
					HaveKeyWithValue("delete_automated_backups", true),
					HaveKeyWithValue("maintenance_day", BeNil()),
					HaveKeyWithValue("maintenance_start_hour", BeNil()),
					HaveKeyWithValue("maintenance_start_min", BeNil()),
					HaveKeyWithValue("maintenance_end_hour", BeNil()),
					HaveKeyWithValue("maintenance_end_min", BeNil()),
					HaveKeyWithValue("allow_major_version_upgrade", false),
					HaveKeyWithValue("auto_minor_version_upgrade", false),
					HaveKeyWithValue("require_ssl", true),
					HaveKeyWithValue("character_set_name", BeNil()),
					HaveKeyWithValue("performance_insights_enabled", false),
					HaveKeyWithValue("performance_insights_kms_key_id", ""),
					HaveKeyWithValue("performance_insights_retention_period", BeNumerically("==", 7)),
					HaveKeyWithValue("enable_export_agent_logs", true),
					HaveKeyWithValue("cloudwatch_agent_log_group_retention_in_days", BeNumerically("==", 1)),
					HaveKeyWithValue("enable_export_error_logs", true),
					HaveKeyWithValue("cloudwatch_error_log_group_retention_in_days", BeNumerically("==", 1)),
					HaveKeyWithValue("cloudwatch_log_groups_kms_key_id", "arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx"),
					HaveKeyWithValue("multi_az", BeTrue()),
					HaveKeyWithValue("use_managed_admin_password", false),
					HaveKeyWithValue("rotate_admin_password_after", float64(7)),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(msSQLServiceName, customMSSQLPlan["name"].(string), buildProperties(requiredProperties(), map[string]any{
				"use_managed_admin_password":  true,
				"rotate_admin_password_after": 365,
			}))
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("use_managed_admin_password", true),
					HaveKeyWithValue("rotate_admin_password_after", float64(365)),
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
			Entry("should prevent updating", "character_set_name", "no-matter-its-value"),
		)

		DescribeTable("should allow unsetting properties flagged as `nullable` by explicitly updating their value to be `nil`",
			func(prop string, initValue any) {
				err := broker.Update(instanceID, msSQLServiceName, customMSSQLPlan["name"].(string), map[string]any{prop: initValue})
				Expect(err).NotTo(HaveOccurred())
				Expect(nthTerraformInvocationVars(mockTerraform, 1)).To(HaveKeyWithValue(prop, initValue))

				err = broker.Update(instanceID, msSQLServiceName, customMSSQLPlan["name"].(string), map[string]any{prop: nil})
				Expect(err).NotTo(HaveOccurred())
				Expect(nthTerraformInvocationVars(mockTerraform, 2)).To(HaveKeyWithValue(prop, BeNil()))
			},
			Entry("max_allocated_storage is nullable", "max_allocated_storage", float64(987)),
			Entry("backup_window is nullable", "backup_window", "00:00-00:00"),
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
			Entry("update engine", "engine", "no-matter-what-engine"),
			Entry("update storage_encrypted", "storage_encrypted", true),
			Entry("update kms_key_id", "kms_key_id", "no-matter-what-key"),
			Entry("update db_name", "db_name", "no-matter-what-name"),
			Entry("update instance_name", "instance_name", "no-matter-what-name"),
			Entry("update region", "region", "no-matter-what-region"),
			Entry("rds_vpc_security_group_ids", "rds_vpc_security_group_ids", "group3"),
			Entry("character_set_name", "character_set_name", "no-matter-what-character-set"),
		)

		DescribeTable("should allow updating properties",
			func(prop string, value any) {
				err := broker.Update(instanceID, msSQLServiceName, customMySQLPlan["name"].(string), map[string]any{prop: value})

				Expect(err).NotTo(HaveOccurred())
				Expect(nthTerraformInvocationVars(mockTerraform, 1)).To(HaveKeyWithValue(prop, value))
			},
			Entry("update storage_type", "storage_type", "gp2"),
			Entry("update iops", "iops", float64(1500)),
			Entry("update deletion_protection", "deletion_protection", true),
			Entry("update publicly_accessible", "publicly_accessible", true),
			Entry("update backup_retention_period", "backup_retention_period", float64(2)),
			Entry("update backup_window", "backup_window", "01:02-03:04"),
			Entry("update copy_tags_to_snapshot", "copy_tags_to_snapshot", false),
			Entry("update delete_automated_backups", "delete_automated_backups", false),
			Entry("update allow_major_version_upgrade", "allow_major_version_upgrade", false),
			Entry("update auto_minor_version_upgrade", "auto_minor_version_upgrade", false),
			Entry("update require_ssl", "require_ssl", false),
			Entry("update enable_export_agent_logs", "enable_export_agent_logs", true),
			Entry("update enable_export_error_logs", "enable_export_error_logs", true),
			Entry("update cloudwatch_log_groups_kms_key_id", "cloudwatch_log_groups_kms_key_id", "arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx"),
		)
	})
})
