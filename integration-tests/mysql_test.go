package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mySQLServiceID                  = "fa22af0f-3637-4a36-b8a7-cfc61168a3e0"
	mySQLServiceName                = "csb-aws-mysql"
	mySQLServiceDescription         = "CSB Amazon RDS for MySQL"
	mySQLServiceDisplayName         = "CSB Amazon RDS for MySQL"
	mySQLServiceDocumentationURL    = "https://docs.vmware.com/en/Tanzu-Cloud-Service-Broker-for-AWS/1.2/csb-aws/GUID-reference-aws-mysql.html"
	mySQLServiceSupportURL          = "https://aws.amazon.com/rds/mysql/resources/?nc=sn&loc=5"
	mySQLServiceProviderDisplayName = "VMware"
	mySQLCustomPlanName             = "custom-sample"
	mySQLCustomPlanID               = "c2ae1820-8c1a-4cf7-90cf-8340ba5aa0bf"
)

var customMySQLPlans = []map[string]any{
	customMySQLPlan,
}

var customMySQLPlan = map[string]any{
	"name":          mySQLCustomPlanName,
	"id":            mySQLCustomPlanID,
	"description":   "Default MySQL plan",
	"mysql_version": 8,
	"cores":         4,
	"storage_gb":    100,
	"metadata": map[string]any{
		"displayName": "custom-sample",
	},
}

var _ = Describe("MySQL", Label("MySQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish AWS MySQL in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mySQLServiceName)
		Expect(service.ID).To(Equal(mySQLServiceID))
		Expect(service.Description).To(Equal(mySQLServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "mysql"))
		Expect(service.Metadata.DisplayName).To(Equal(mySQLServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(mySQLServiceDocumentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(mySQLServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(mySQLServiceProviderDisplayName))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal(mySQLCustomPlanName),
					ID:   Equal(mySQLCustomPlanID),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(mySQLServiceName, "custom-sample", params)

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
			instanceID, err := broker.Provision(mySQLServiceName, customMySQLPlan["name"].(string), nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("storage_gb", float64(100)),
					HaveKeyWithValue("storage_type", "io1"),
					HaveKeyWithValue("iops", float64(3000)),
					HaveKeyWithValue("storage_autoscale", true),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(250)),
					HaveKeyWithValue("storage_encrypted", true),
					HaveKeyWithValue("kms_key_id", ""),
					HaveKeyWithValue("parameter_group_name", ""),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-mysql-%s", instanceID)),
					HaveKeyWithValue("db_name", "vsbdb"),
					HaveKeyWithValue("publicly_accessible", false),
					HaveKeyWithValue("region", fakeRegion),
					HaveKeyWithValue("multi_az", true),
					HaveKeyWithValue("instance_class", ""),
					HaveKeyWithValue("rds_subnet_group", ""),
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
					HaveKeyWithValue("option_group_name", ""),
					HaveKeyWithValue("monitoring_interval", float64(0)),
					HaveKeyWithValue("monitoring_role_arn", ""),
					HaveKeyWithValue("performance_insights_enabled", false),
					HaveKeyWithValue("performance_insights_kms_key_id", ""),
					HaveKeyWithValue("performance_insights_retention_period", float64(7)),
					HaveKeyWithValue("enable_audit_logging", false),
					HaveKeyWithValue("cloudwatch_log_group_kms_key_id", ""),
					HaveKeyWithValue("cloudwatch_log_group_retention_in_days", float64(30)),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(mySQLServiceName, customMySQLPlan["name"].(string), map[string]any{
				"storage_type":                           "gp2",
				"storage_autoscale":                      true,
				"storage_autoscale_limit_gb":             float64(150),
				"storage_encrypted":                      true,
				"kms_key_id":                             "arn:aws:xxxx",
				"parameter_group_name":                   "fake-parameter-group",
				"instance_name":                          "csb-mysql-fake-name",
				"db_name":                                "fake-db-name",
				"publicly_accessible":                    true,
				"region":                                 "africa-north-4",
				"multi_az":                               true,
				"instance_class":                         "",
				"rds_subnet_group":                       "",
				"rds_vpc_security_group_ids":             "group1,group2",
				"allow_major_version_upgrade":            false,
				"auto_minor_version_upgrade":             false,
				"maintenance_day":                        "Mon",
				"maintenance_start_hour":                 "03",
				"maintenance_start_min":                  "45",
				"maintenance_end_hour":                   "10",
				"maintenance_end_min":                    "15",
				"deletion_protection":                    true,
				"backup_retention_period":                float64(2),
				"backup_window":                          "01:02-03:04",
				"copy_tags_to_snapshot":                  false,
				"delete_automated_backups":               false,
				"option_group_name":                      "option-group-name",
				"monitoring_interval":                    30,
				"monitoring_role_arn":                    "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access",
				"performance_insights_enabled":           true,
				"performance_insights_kms_key_id":        "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa",
				"performance_insights_retention_period":  93,
				"enable_audit_logging":                   true,
				"cloudwatch_log_group_kms_key_id":        "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa",
				"cloudwatch_log_group_retention_in_days": 33,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("engine", "mysql"),
					HaveKeyWithValue("engine_version", "8"),
					HaveKeyWithValue("cores", float64(4)),
					HaveKeyWithValue("storage_gb", float64(100)),
					HaveKeyWithValue("storage_type", "gp2"),
					HaveKeyWithValue("storage_autoscale", true),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(150)),
					HaveKeyWithValue("storage_encrypted", true),
					HaveKeyWithValue("kms_key_id", "arn:aws:xxxx"),
					HaveKeyWithValue("parameter_group_name", "fake-parameter-group"),
					HaveKeyWithValue("instance_name", "csb-mysql-fake-name"),
					HaveKeyWithValue("db_name", "fake-db-name"),
					HaveKeyWithValue("publicly_accessible", true),
					HaveKeyWithValue("region", "africa-north-4"),
					HaveKeyWithValue("multi_az", true),
					HaveKeyWithValue("instance_class", ""),
					HaveKeyWithValue("rds_subnet_group", ""),
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
					HaveKeyWithValue("option_group_name", "option-group-name"),
					HaveKeyWithValue("monitoring_interval", float64(30)),
					HaveKeyWithValue("monitoring_role_arn", "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access"),
					HaveKeyWithValue("performance_insights_enabled", true),
					HaveKeyWithValue("performance_insights_kms_key_id", "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
					HaveKeyWithValue("performance_insights_retention_period", float64(93)),
					HaveKeyWithValue("enable_audit_logging", true),
					HaveKeyWithValue("cloudwatch_log_group_kms_key_id", "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
					HaveKeyWithValue("cloudwatch_log_group_retention_in_days", float64(33)),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(mySQLServiceName, customMySQLPlan["name"].(string), nil)

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance",
			func(prop string, value any) {
				err := broker.Update(instanceID, mySQLServiceName, customMySQLPlan["name"].(string), map[string]any{prop: value})

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
			Entry("update kms_key_id", "kms_key_id", "no-matter-what-key"),
			Entry("update storage_encrypted", "storage_encrypted", true),
			Entry("rds_vpc_security_group_ids", "rds_vpc_security_group_ids", "group3"),
		)

		DescribeTable("should allow updating properties",
			func(prop string, value any) {
				err := broker.Update(instanceID, mySQLServiceName, customMySQLPlan["name"].(string), map[string]any{prop: value})

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update storage_type", "storage_type", "gp2"),
			Entry("update iops", "iops", 1500),
			Entry("update storage_autoscale", "storage_autoscale", true),
			Entry("update storage_autoscale_limit_gb", "storage_autoscale_limit_gb", 2),
			Entry("update deletion_protection", "deletion_protection", false),
			Entry("update backup_retention_period", "backup_retention_period", float64(2)),
			Entry("update backup_window", "backup_window", "01:02-03:04"),
			Entry("update copy_tags_to_snapshot", "copy_tags_to_snapshot", false),
			Entry("update delete_automated_backups", "delete_automated_backups", false),
			Entry("update option_group_name", "option_group_name", "option-group-name"),
			Entry("update monitoring_interval", "monitoring_interval", 0),
			Entry("update monitoring_role_arn", "monitoring_role_arn", ""),
			Entry("update performance_insights_enabled", "performance_insights_enabled", true),
			Entry("update performance_insights_kms_key_id", "performance_insights_kms_key_id", "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
			Entry("update performance_insights_retention_period", "performance_insights_retention_period", 31),
		)
	})
})
