package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	auroraPostgreSQLServiceID                  = "36203e40-2945-11ed-8980-eb81bd131a02"
	auroraPostgreSQLServiceName                = "csb-aws-aurora-postgresql"
	auroraPostgreSQLServiceDescription         = "Amazon Aurora for PostgreSQL"
	auroraPostgreSQLServiceDisplayName         = "Amazon Aurora for PostgreSQL"
	auroraPostgreSQLServiceSupportURL          = "https://aws.amazon.com/rds/aurora/"
	auroraPostgreSQLServiceProviderDisplayName = "VMware"
	auroraPostgreSQLCustomPlanName             = "custom-sample"
	auroraPostgreSQLCustomPlanID               = "d20c5cf2-29e1-11ed-93da-1f3a67a06903"
)

var customAuroraPostgresPlans = []map[string]any{
	customAuroraPostgresPlan,
}

var customAuroraPostgresPlan = map[string]any{
	"name":        auroraPostgreSQLCustomPlanName,
	"id":          auroraPostgreSQLCustomPlanID,
	"description": "Default Aurora Postgres plan",
	"metadata": map[string]any{
		"displayName": "custom-sample",
	},
}

var _ = Describe("Aurora PostgreSQL", Label("aurora-postgresql"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish Aurora Postgres in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, auroraPostgreSQLServiceName)
		Expect(service.ID).To(Equal(auroraPostgreSQLServiceID))
		Expect(service.Description).To(Equal(auroraPostgreSQLServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "postgres", "postgresql", "aurora"))
		Expect(service.Metadata.DisplayName).To(Equal(auroraPostgreSQLServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(documentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(auroraPostgreSQLServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(auroraPostgreSQLServiceProviderDisplayName))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(auroraPostgreSQLCustomPlanID),
					Name: Equal(auroraPostgreSQLCustomPlanName),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		DescribeTable("should check property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(auroraPostgreSQLServiceName, "custom-sample", params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-cluster.html#options
				"instance name will be used as db-cluster-identifier so must contain from 1 to 63 letters, numbers or hyphens",
				map[string]any{"instance_name": stringOfLen(64)},
				"instance_name: String length must be less than or equal to 63",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-cluster.html#options
				"instance name will be used as db-cluster-identifier so the first character must be a letter",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-zA-Z](-?[a-zA-Z0-9])*$'",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-cluster.html#options
				"instance name will be used as db-cluster-identifier so it cannot end with a hyphen",
				map[string]any{"instance_name": "aaaaa-"},
				"instance_name: Does not match pattern '^[a-zA-Z](-?[a-zA-Z0-9])*$'",
			),
			Entry(
				// https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-cluster.html#options
				"instance name will be used as db-cluster-identifier so it cannot contain two consecutive hyphens",
				map[string]any{"instance_name": "aa--aaa"},
				"instance_name: Does not match pattern '^[a-zA-Z](-?[a-zA-Z0-9])*$'",
			),
			Entry(
				"database name maximum length is 64 characters",
				map[string]any{"db_name": stringOfLen(65)},
				"db_name: String length must be less than or equal to 64",
			),
			Entry(
				"database name invalid characters",
				map[string]any{"db_name": ".aaaaa"},
				"db_name: Does not match pattern '^[a-z][a-z0-9_]+$'",
			),
			Entry(
				"monitoring_interval maximum value is 60",
				map[string]any{"monitoring_interval": 61},
				"monitoring_interval: Must be less than or equal to 60",
			),
			Entry(
				"monitoring_interval minimum value is 0",
				map[string]any{"monitoring_interval": -1},
				"monitoring_interval: Must be greater than or equal to 0",
			),
			Entry(
				"performance_insights_retention_period minimum value is 7",
				map[string]any{"performance_insights_retention_period": 1},
				"performance_insights_retention_period: Must be greater than or equal to 7",
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(auroraPostgreSQLServiceName, "custom-sample", map[string]any{
				"engine_version": "13.7",
				"instance_class": "db.r5.large",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-aurorapg-%s", instanceID)),
					HaveKeyWithValue("cluster_instances", BeNumerically("==", 3)),
					HaveKeyWithValue("db_name", "csbdb"),
					HaveKeyWithValue("region", fakeRegion),
					HaveKeyWithValue("allow_major_version_upgrade", BeTrue()),
					HaveKeyWithValue("auto_minor_version_upgrade", BeTrue()),
					HaveKeyWithValue("rds_vpc_security_group_ids", BeEmpty()),
					HaveKeyWithValue("rds_subnet_group", BeEmpty()),
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{
						"pcf-instance-id": Equal(instanceID),
						"key1":            Equal("value1"),
						"key2":            Equal("value2"),
					})),
					HaveKeyWithValue("deletion_protection", BeFalse()),
					HaveKeyWithValue("engine_version", "13.7"),
					HaveKeyWithValue("monitoring_interval", BeNumerically("==", 0)),
					HaveKeyWithValue("monitoring_role_arn", ""),
					HaveKeyWithValue("performance_insights_enabled", false),
					HaveKeyWithValue("performance_insights_kms_key_id", ""),
					HaveKeyWithValue("performance_insights_retention_period", BeNumerically("==", 7)),
					HaveKeyWithValue("storage_encrypted", true),
					HaveKeyWithValue("kms_key_id", ""),
					HaveKeyWithValue("instance_class", "db.r5.large"),
					HaveKeyWithValue("preferred_maintenance_day", BeNil()),
					HaveKeyWithValue("preferred_maintenance_start_hour", BeNil()),
					HaveKeyWithValue("preferred_maintenance_start_min", BeNil()),
					HaveKeyWithValue("preferred_maintenance_end_hour", BeNil()),
					HaveKeyWithValue("preferred_maintenance_end_min", BeNil()),
				))
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(auroraPostgreSQLServiceName, "custom-sample", map[string]any{
				"instance_name":                         "csb-aurora-postgres-fake-name",
				"db_name":                               "fakedbname",
				"region":                                "africa-north-4",
				"cluster_instances":                     12,
				"serverless_min_capacity":               0.2,
				"serverless_max_capacity":               100,
				"allow_major_version_upgrade":           false,
				"auto_minor_version_upgrade":            false,
				"rds_vpc_security_group_ids":            "group1,group2",
				"rds_subnet_group":                      "some-other-subnet",
				"deletion_protection":                   true,
				"engine_version":                        "8.0.postgresql_aurora.3.02.0",
				"monitoring_interval":                   30,
				"monitoring_role_arn":                   "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access",
				"performance_insights_enabled":          true,
				"performance_insights_kms_key_id":       "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa",
				"performance_insights_retention_period": 93,
				"storage_encrypted":                     false,
				"kms_key_id":                            "arn:aws:kms:us-south-10:123456789012:key/face1945-7581-4bf6-b311-39594be3dce5",
				"instance_class":                        "db.r5.large",
				"preferred_maintenance_day":             "Mon",
				"preferred_maintenance_start_hour":      "03",
				"preferred_maintenance_start_min":       "45",
				"preferred_maintenance_end_hour":        "10",
				"preferred_maintenance_end_min":         "15",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("instance_name", "csb-aurora-postgres-fake-name"),
					HaveKeyWithValue("db_name", "fakedbname"),
					HaveKeyWithValue("region", "africa-north-4"),
					HaveKeyWithValue("cluster_instances", BeNumerically("==", 12)),
					HaveKeyWithValue("serverless_min_capacity", BeNumerically("==", 0.2)),
					HaveKeyWithValue("serverless_max_capacity", BeNumerically("==", 100)),
					HaveKeyWithValue("engine_version", "8.0.postgresql_aurora.3.02.0"),
					HaveKeyWithValue("allow_major_version_upgrade", false),
					HaveKeyWithValue("auto_minor_version_upgrade", false),
					HaveKeyWithValue("rds_vpc_security_group_ids", "group1,group2"),
					HaveKeyWithValue("rds_subnet_group", "some-other-subnet"),
					HaveKeyWithValue("deletion_protection", true),
					HaveKeyWithValue("monitoring_interval", BeNumerically("==", 30)),
					HaveKeyWithValue("monitoring_role_arn", "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access"),
					HaveKeyWithValue("performance_insights_enabled", true),
					HaveKeyWithValue("performance_insights_kms_key_id", "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
					HaveKeyWithValue("performance_insights_retention_period", BeNumerically("==", 93)),
					HaveKeyWithValue("storage_encrypted", false),
					HaveKeyWithValue("kms_key_id", "arn:aws:kms:us-south-10:123456789012:key/face1945-7581-4bf6-b311-39594be3dce5"),
					HaveKeyWithValue("instance_class", "db.r5.large"),
					HaveKeyWithValue("preferred_maintenance_day", "Mon"),
					HaveKeyWithValue("preferred_maintenance_start_hour", "03"),
					HaveKeyWithValue("preferred_maintenance_start_min", "45"),
					HaveKeyWithValue("preferred_maintenance_end_hour", "10"),
					HaveKeyWithValue("preferred_maintenance_end_min", "15"),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(auroraPostgreSQLServiceName, "custom-sample", map[string]any{
				"engine_version": "13.7",
				"instance_class": "db.r5.large",
			})

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable(
			"preventing updates with `prohibit_update` as it can force resource replacement or re-creation",
			func(prop string, value any) {
				err := broker.Update(instanceID, auroraPostgreSQLServiceName, "custom-sample", map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("region", "region", "no-matter-what-region"),
			Entry("instance_name", "instance_name", "marmaduke"),
			Entry("db_name", "db_name", "someNewName"),
			Entry("rds_subnet_group", "rds_subnet_group", "some-new-subnet-name"),
			Entry("rds_vpc_security_group_ids", "rds_vpc_security_group_ids", "group3"),
			Entry("storage_encrypted", "storage_encrypted", false),
			Entry("kms_key_id", "kms_key_id", "arn:aws:kms:eu-north-42:741085209630:key/a2c0ffee-cab0-4617-a28e-cabba9e06193"),
		)

		DescribeTable(
			"allowed updates",
			func(prop string, value any) {
				Expect(broker.Update(instanceID, auroraPostgreSQLServiceName, "custom-sample", map[string]any{prop: value})).To(Succeed())
			},
			Entry("cluster_instances", "cluster_instances", 11),
			Entry("serverless_min_capacity", "serverless_min_capacity", 1),
			Entry("serverless_max_capacity", "serverless_max_capacity", 30),
			Entry("allow_major_version_upgrade", "allow_major_version_upgrade", false),
			Entry("auto_minor_version_upgrade", "auto_minor_version_upgrade", false),
			Entry("deletion_protection", "deletion_protection", true),
			Entry("engine_version", "engine_version", "8.0.postgresql_aurora.3.02.0"),
			Entry("update monitoring_interval", "monitoring_interval", 0),
			Entry("update monitoring_role_arn", "monitoring_role_arn", ""),
			Entry("update performance_insights_enabled", "performance_insights_enabled", true),
			Entry("update performance_insights_kms_key_id", "performance_insights_kms_key_id", "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
			Entry("update performance_insights_retention_period", "performance_insights_retention_period", 31),
			Entry("update instance_class", "instance_class", "db.r5.large"),
		)
	})
})
