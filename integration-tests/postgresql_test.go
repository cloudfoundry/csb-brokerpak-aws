package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	postgreSQLServiceID                  = "fa6334bc-5314-4b63-8a74-c0e4b638c950"
	postgreSQLServiceName                = "csb-aws-postgresql"
	postgreSQLServiceDescription         = "CSB Amazon RDS for PostgreSQL"
	postgreSQLServiceDisplayName         = "CSB Amazon RDS for PostgreSQL"
	postgreSQLServiceSupportURL          = "https://aws.amazon.com/rds/postgresql/"
	postgreSQLServiceProviderDisplayName = "VMware"
	postgreSQLCustomPlanName             = "custom-sample"
	postgreSQLCustomPlanID               = "de7dbcee-1c8d-11ed-9904-5f435c1e2316"
)

var customPostgresPlans = []map[string]any{
	customPostgresPlan,
}

var customPostgresPlan = map[string]any{
	"name":        postgreSQLCustomPlanName,
	"id":          postgreSQLCustomPlanID,
	"description": "Default Postgres plan",
	"metadata": map[string]any{
		"displayName": "custom-sample",
	},
	"instance_class":   "db.m6i.large",
	"postgres_version": "14.2",
	"storage_gb":       100,
}

var _ = Describe("Postgresql", Label("Postgresql"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish AWS postgres in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, postgreSQLServiceName)
		Expect(service.ID).To(Equal(postgreSQLServiceID))
		Expect(service.Description).To(Equal(postgreSQLServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "postgres", "postgresql"))
		Expect(service.Metadata.DisplayName).To(Equal(postgreSQLServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(documentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(postgreSQLServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(postgreSQLServiceProviderDisplayName))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal(postgreSQLCustomPlanName),
					ID:   Equal(postgreSQLCustomPlanID),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(postgreSQLServiceName, "custom-sample", params)

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
				"instance name invalid characters",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"database name maximum length is 98 characters",
				map[string]any{"db_name": stringOfLen(99)},
				"db_name: String length must be less than or equal to 64",
			),
			Entry(
				"performance_insights_retention_period minimum value is 7",
				map[string]any{"performance_insights_retention_period": 1},
				"performance_insights_retention_period: Must be greater than or equal to 7",
			),
			Entry(
				"cloudwatch_postgresql_log_group_retention_in_days minimum value is 0",
				map[string]any{"cloudwatch_postgresql_log_group_retention_in_days": -1},
				"cloudwatch_postgresql_log_group_retention_in_days: Must be greater than or equal to 0",
			),
			Entry(
				"cloudwatch_postgresql_log_group_retention_in_days maximum value is 3653",
				map[string]any{"cloudwatch_postgresql_log_group_retention_in_days": 3654},
				"cloudwatch_postgresql_log_group_retention_in_days: Must be less than or equal to 3653",
			),
			Entry(
				"cloudwatch_upgrade_log_group_retention_in_days minimum value is 0",
				map[string]any{"cloudwatch_upgrade_log_group_retention_in_days": -1},
				"cloudwatch_upgrade_log_group_retention_in_days: Must be greater than or equal to 0",
			),
			Entry(
				"cloudwatch_upgrade_log_group_retention_in_days maximum value is 3653",
				map[string]any{"cloudwatch_upgrade_log_group_retention_in_days": 3654},
				"cloudwatch_upgrade_log_group_retention_in_days: Must be less than or equal to 3653",
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(postgreSQLServiceName, "custom-sample", nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{
						"pcf-instance-id": Equal(instanceID),
						"key1":            Equal("value1"),
						"key2":            Equal("value2"),
					})),
					HaveKeyWithValue("postgres_version", "14.2"),
					HaveKeyWithValue("storage_gb", float64(100)),
					HaveKeyWithValue("storage_type", "io1"),
					HaveKeyWithValue("iops", float64(3000)),
					HaveKeyWithValue("require_ssl", false),
					HaveKeyWithValue("provider_verify_certificate", true),
					HaveKeyWithValue("storage_autoscale", false),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(0)),
					HaveKeyWithValue("parameter_group_name", ""),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-postgresql-%s", instanceID)),
					HaveKeyWithValue("db_name", "vsbdb"),
					HaveKeyWithValue("publicly_accessible", false),
					HaveKeyWithValue("region", fakeRegion),
					HaveKeyWithValue("storage_encrypted", false),
					HaveKeyWithValue("kms_key_id", ""),
					HaveKeyWithValue("multi_az", false),
					HaveKeyWithValue("rds_vpc_security_group_ids", ""),
					HaveKeyWithValue("allow_major_version_upgrade", true),
					HaveKeyWithValue("auto_minor_version_upgrade", true),
					HaveKeyWithValue("maintenance_day", BeNil()),
					HaveKeyWithValue("maintenance_start_hour", BeNil()),
					HaveKeyWithValue("maintenance_start_min", BeNil()),
					HaveKeyWithValue("maintenance_end_hour", BeNil()),
					HaveKeyWithValue("maintenance_end_min", BeNil()),
					HaveKeyWithValue("deletion_protection", false),
					HaveKeyWithValue("backup_retention_period", float64(7)),
					HaveKeyWithValue("backup_window", BeNil()),
					HaveKeyWithValue("copy_tags_to_snapshot", true),
					HaveKeyWithValue("delete_automated_backups", true),
					HaveKeyWithValue("monitoring_interval", float64(0)),
					HaveKeyWithValue("monitoring_role_arn", ""),
					HaveKeyWithValue("performance_insights_enabled", false),
					HaveKeyWithValue("performance_insights_kms_key_id", ""),
					HaveKeyWithValue("performance_insights_retention_period", float64(7)),
					HaveKeyWithValue("enable_export_postgresql_logs", false),
					HaveKeyWithValue("cloudwatch_postgresql_log_group_retention_in_days", BeNumerically("==", 30)),
					HaveKeyWithValue("enable_export_upgrade_logs", false),
					HaveKeyWithValue("cloudwatch_upgrade_log_group_retention_in_days", BeNumerically("==", 30)),
					HaveKeyWithValue("cloudwatch_log_groups_kms_key_id", ""),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(postgreSQLServiceName, "custom-sample", map[string]any{
				"require_ssl":                           true,
				"storage_type":                          "gp2",
				"provider_verify_certificate":           false,
				"storage_autoscale":                     true,
				"storage_autoscale_limit_gb":            float64(150),
				"parameter_group_name":                  "flopsy",
				"instance_name":                         "csb-postgresql-mopsy",
				"db_name":                               "cottontail",
				"publicly_accessible":                   true,
				"region":                                "africa-north-4",
				"storage_encrypted":                     true,
				"kms_key_id":                            "arn:aws:xxxx",
				"multi_az":                              true,
				"rds_vpc_security_group_ids":            "group1,group2",
				"allow_major_version_upgrade":           false,
				"auto_minor_version_upgrade":            false,
				"maintenance_day":                       "Mon",
				"maintenance_start_hour":                "03",
				"maintenance_start_min":                 "45",
				"maintenance_end_hour":                  "10",
				"maintenance_end_min":                   "15",
				"deletion_protection":                   true,
				"backup_retention_period":               float64(2),
				"backup_window":                         "01:02-03:04",
				"copy_tags_to_snapshot":                 false,
				"delete_automated_backups":              false,
				"monitoring_interval":                   30,
				"monitoring_role_arn":                   "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access",
				"performance_insights_enabled":          true,
				"performance_insights_kms_key_id":       "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa",
				"performance_insights_retention_period": 93,
				"enable_export_postgresql_logs":         true,
				"cloudwatch_postgresql_log_group_retention_in_days": 1,
				"enable_export_upgrade_logs":                        true,
				"cloudwatch_upgrade_log_group_retention_in_days":    1,
				"cloudwatch_log_groups_kms_key_id":                  "arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("require_ssl", true),
					HaveKeyWithValue("storage_type", "gp2"),
					HaveKeyWithValue("provider_verify_certificate", false),
					HaveKeyWithValue("storage_autoscale", true),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(150)),
					HaveKeyWithValue("parameter_group_name", "flopsy"),
					HaveKeyWithValue("instance_name", "csb-postgresql-mopsy"),
					HaveKeyWithValue("db_name", "cottontail"),
					HaveKeyWithValue("publicly_accessible", true),
					HaveKeyWithValue("region", "africa-north-4"),
					HaveKeyWithValue("storage_encrypted", true),
					HaveKeyWithValue("kms_key_id", "arn:aws:xxxx"),
					HaveKeyWithValue("multi_az", true),
					HaveKeyWithValue("rds_vpc_security_group_ids", "group1,group2"),
					HaveKeyWithValue("allow_major_version_upgrade", false),
					HaveKeyWithValue("auto_minor_version_upgrade", false),
					HaveKeyWithValue("maintenance_day", "Mon"),
					HaveKeyWithValue("maintenance_start_hour", "03"),
					HaveKeyWithValue("maintenance_start_min", "45"),
					HaveKeyWithValue("maintenance_end_hour", "10"),
					HaveKeyWithValue("maintenance_end_min", "15"),
					HaveKeyWithValue("deletion_protection", true),
					HaveKeyWithValue("backup_retention_period", float64(2)),
					HaveKeyWithValue("backup_window", "01:02-03:04"),
					HaveKeyWithValue("copy_tags_to_snapshot", false),
					HaveKeyWithValue("delete_automated_backups", false),
					HaveKeyWithValue("monitoring_interval", float64(30)),
					HaveKeyWithValue("monitoring_role_arn", "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access"),
					HaveKeyWithValue("performance_insights_enabled", true),
					HaveKeyWithValue("performance_insights_kms_key_id", "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
					HaveKeyWithValue("performance_insights_retention_period", float64(93)),
					HaveKeyWithValue("enable_export_postgresql_logs", true),
					HaveKeyWithValue("cloudwatch_postgresql_log_group_retention_in_days", BeNumerically("==", 1)),
					HaveKeyWithValue("enable_export_upgrade_logs", true),
					HaveKeyWithValue("cloudwatch_upgrade_log_group_retention_in_days", BeNumerically("==", 1)),
					HaveKeyWithValue("cloudwatch_log_groups_kms_key_id", "arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx"),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(postgreSQLServiceName, "custom-sample", nil)

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance",
			func(prop string, value any) {
				err := broker.Update(instanceID, postgreSQLServiceName, "custom-sample", map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update region", "region", "no-matter-what-region"),
			Entry("update kms_key_id", "kms_key_id", "no-matter-what-key"),
			Entry("update db_name", "db_name", "no-matter-what-name"),
			Entry("update storage_encrypted", "storage_encrypted", true),
			Entry("rds_vpc_security_group_ids", "rds_vpc_security_group_ids", "group3"),
		)

		DescribeTable(
			"some allowed updates",
			func(prop string, value any) {
				err := broker.Update(instanceID, postgreSQLServiceName, "custom-sample", map[string]any{prop: value})

				Expect(err).NotTo(HaveOccurred())
			},
			Entry(nil, "require_ssl", true),
			Entry(nil, "storage_type", "gp2"),
			Entry(nil, "provider_verify_certificate", false),
			Entry(nil, "deletion_protection", true),
			Entry(nil, "monitoring_interval", 0),
			Entry(nil, "monitoring_role_arn", ""),
			Entry(nil, "backup_retention_period", float64(2)),
			Entry(nil, "backup_window", "01:02-03:04"),
			Entry(nil, "copy_tags_to_snapshot", false),
			Entry(nil, "delete_automated_backups", false),
			Entry(nil, "monitoring_interval", 30),
			Entry(nil, "monitoring_role_arn", "arn:aws:iam::649758297924:role/enhanced_monitoring_access"),
			Entry(nil, "performance_insights_enabled", true),
			Entry(nil, "performance_insights_kms_key_id", "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
			Entry(nil, "performance_insights_retention_period", 31),
			Entry(nil, "enable_export_postgresql_logs", true),
			Entry(nil, "enable_export_upgrade_logs", true),
			Entry(nil, "cloudwatch_log_groups_kms_key_id", "arn:aws:kms:us-west-2:xxxxxxxxxxxx:key/xxxxxxxx-80b9-4afd-98c0-xxxxxxxxxxxx"),
		)
	})
})
