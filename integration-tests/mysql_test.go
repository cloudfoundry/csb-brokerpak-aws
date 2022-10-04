package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var customMySQLPlans = []map[string]any{
	customMySQLPlan,
}

var customMySQLPlan = map[string]any{
	"name":          "custom-sample",
	"id":            "c2ae1820-8c1a-4cf7-90cf-8340ba5aa0bf",
	"description":   "Beta - Default MySQL plan",
	"mysql_version": 8,
	"cores":         4,
	"storage_gb":    100,
	"metadata": map[string]any{
		"displayName": "custom-sample (Beta)",
	},
}

var _ = Describe("MySQL", Label("MySQL"), func() {
	const serviceName = "csb-aws-mysql"

	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish AWS MySQL in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, serviceName)
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf("aws", "mysql", "beta"))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal("small")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("medium")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("large")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("custom-sample")}),
			),
		)

		Expect(service.Plans).To(
			HaveEach(
				MatchFields(IgnoreExtras, Fields{
					"Description": HavePrefix("Beta -"),
					"Metadata":    PointTo(MatchFields(IgnoreExtras, Fields{"DisplayName": HaveSuffix("(Beta)")})),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(serviceName, "custom-sample", params)

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
			instanceID, err := broker.Provision(serviceName, customMySQLPlan["name"].(string), nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("use_tls", true),
					HaveKeyWithValue("storage_gb", float64(100)),
					HaveKeyWithValue("storage_type", "io1"),
					HaveKeyWithValue("iops", float64(3000)),
					HaveKeyWithValue("storage_autoscale", false),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(0)),
					HaveKeyWithValue("storage_encrypted", false),
					HaveKeyWithValue("parameter_group_name", ""),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-mysql-%s", instanceID)),
					HaveKeyWithValue("db_name", "vsbdb"),
					HaveKeyWithValue("publicly_accessible", false),
					HaveKeyWithValue("region", "us-west-2"),
					HaveKeyWithValue("multi_az", false),
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
					HaveKeyWithValue("option_group_name", BeNil()),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(serviceName, customMySQLPlan["name"].(string), map[string]any{
				"use_tls":                     false,
				"storage_type":                "gp2",
				"storage_autoscale":           true,
				"storage_autoscale_limit_gb":  float64(150),
				"storage_encrypted":           true,
				"parameter_group_name":        "fake-parameter-group",
				"instance_name":               "csb-mysql-fake-name",
				"db_name":                     "fake-db-name",
				"publicly_accessible":         true,
				"region":                      "africa-north-4",
				"multi_az":                    true,
				"instance_class":              "",
				"rds_subnet_group":            "",
				"rds_vpc_security_group_ids":  "",
				"allow_major_version_upgrade": false,
				"auto_minor_version_upgrade":  false,
				"maintenance_day":             "Mon",
				"maintenance_start_hour":      "03",
				"maintenance_start_min":       "45",
				"maintenance_end_hour":        "10",
				"maintenance_end_min":         "15",
				"deletion_protection":         true,
				"backup_retention_period":     float64(2),
				"backup_window":               "01:02-03:04",
				"copy_tags_to_snapshot":       false,
				"delete_automated_backups":    false,
				"option_group_name":           "option-group-name",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("engine", "mysql"),
					HaveKeyWithValue("engine_version", "8"),
					HaveKeyWithValue("cores", float64(4)),
					HaveKeyWithValue("storage_gb", float64(100)),
					HaveKeyWithValue("use_tls", false),
					HaveKeyWithValue("storage_type", "gp2"),
					HaveKeyWithValue("storage_autoscale", true),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(150)),
					HaveKeyWithValue("storage_encrypted", true),
					HaveKeyWithValue("parameter_group_name", "fake-parameter-group"),
					HaveKeyWithValue("instance_name", "csb-mysql-fake-name"),
					HaveKeyWithValue("db_name", "fake-db-name"),
					HaveKeyWithValue("publicly_accessible", true),
					HaveKeyWithValue("region", "africa-north-4"),
					HaveKeyWithValue("multi_az", true),
					HaveKeyWithValue("instance_class", ""),
					HaveKeyWithValue("rds_subnet_group", ""),
					HaveKeyWithValue("rds_vpc_security_group_ids", ""),
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
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(serviceName, "small", nil)

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance and lost data",
			func(params map[string]any) {
				err := broker.Update(instanceID, serviceName, customMySQLPlan["name"].(string), params)

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update region", map[string]any{"region": "no-matter-what-region"}),
			Entry("update db_name", map[string]any{"db_name": "no-matter-what-name"}),
		)

		DescribeTable("should allow updating properties",
			func(params map[string]any) {
				err := broker.Update(instanceID, serviceName, customMySQLPlan["name"].(string), params)

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update use_tls", map[string]any{"use_tls": false}),
			Entry("update storage_type", map[string]any{"storage_type": "gp2"}),
			Entry("update iops", map[string]any{"iops": 1500}),
			Entry("update storage_autoscale", map[string]any{"storage_autoscale": true}),
			Entry("update storage_autoscale_limit_gb", map[string]any{"storage_autoscale_limit_gb": 2}),
			Entry("update storage_encrypted", map[string]any{"storage_encrypted": true}),
			Entry("update deletion_protection", map[string]any{"deletion_protection": false}),
			Entry("update backup_retention_period", map[string]any{"backup_retention_period": float64(2)}),
			Entry("update backup_window", map[string]any{"backup_window": "01:02-03:04"}),
			Entry("update copy_tags_to_snapshot", map[string]any{"copy_tags_to_snapshot": false}),
			Entry("update delete_automated_backups", map[string]any{"delete_automated_backups": false}),
			Entry("update option_group_name", map[string]any{"option_group_name": "option-group-name"}),
		)
	})
})
