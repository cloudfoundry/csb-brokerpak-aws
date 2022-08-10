package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Postgresql", Label("Postgresql"), func() {
	const serviceName = "csb-aws-postgresql"

	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish AWS postgres in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, serviceName)
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf("aws", "postgres", "postgresql", "beta"))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal("small")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("medium")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("large")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("subsume")}),
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
				_, err := broker.Provision(serviceName, "small", params)

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
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(serviceName, "small", nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("cores", float64(2)),
					HaveKeyWithValue("engine_version", "11"),
					HaveKeyWithValue("storage_gb", float64(5)),
					HaveKeyWithValue("subsume", false),
					HaveKeyWithValue("require_ssl", false),
					HaveKeyWithValue("provider_verify_certificate", true),
					HaveKeyWithValue("storage_autoscale", false),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(0)),
					HaveKeyWithValue("parameter_group_name", ""),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-postgresql-%s", instanceID)),
					HaveKeyWithValue("db_name", "vsbdb"),
					HaveKeyWithValue("publicly_accessible", false),
					HaveKeyWithValue("region", "us-west-2"),
					HaveKeyWithValue("storage_encrypted", false),
					HaveKeyWithValue("kms_key_id", ""),
					HaveKeyWithValue("multi_az", false),
					HaveKeyWithValue("allow_major_version_upgrade", true),
					HaveKeyWithValue("auto_minor_version_upgrade", true),
					HaveKeyWithValue("maintenance_window", "Sun:00:00-Sun:00:00"),
					HaveKeyWithValue("deletion_protection", false),
					HaveKeyWithValue("backup_retention_period", float64(7)),
					HaveKeyWithValue("backup_window", "00:00-00:00"),
					HaveKeyWithValue("copy_tags_to_snapshot", true),
					HaveKeyWithValue("delete_automated_backups", true),
					HaveKeyWithValue("monitoring_interval", float64(0)),
					HaveKeyWithValue("monitoring_role_arn", ""),
					HaveKeyWithValue("performance_insights_enabled", false),
					HaveKeyWithValue("performance_insights_kms_key_id", ""),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(serviceName, "small", map[string]any{
				"require_ssl":                     true,
				"provider_verify_certificate":     false,
				"storage_autoscale":               true,
				"storage_autoscale_limit_gb":      float64(10),
				"parameter_group_name":            "flopsy",
				"instance_name":                   "csb-postgresql-mopsy",
				"db_name":                         "cottontail",
				"publicly_accessible":             true,
				"region":                          "africa-north-4",
				"storage_encrypted":               true,
				"kms_key_id":                      "arn:aws:xxxx",
				"multi_az":                        true,
				"allow_major_version_upgrade":     false,
				"auto_minor_version_upgrade":      false,
				"maintenance_day":                 "Mon",
				"maintenance_start_hour":          "03",
				"maintenance_start_min":           "45",
				"maintenance_end_hour":            "10",
				"maintenance_end_min":             "15",
				"deletion_protection":             true,
				"backup_retention_period":         float64(2),
				"backup_window":                   "01:02-03:04",
				"copy_tags_to_snapshot":           false,
				"delete_automated_backups":        false,
				"monitoring_interval":             30,
				"monitoring_role_arn":             "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access",
				"performance_insights_enabled":    true,
				"performance_insights_kms_key_id": "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("require_ssl", true),
					HaveKeyWithValue("provider_verify_certificate", false),
					HaveKeyWithValue("storage_autoscale", true),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(10)),
					HaveKeyWithValue("parameter_group_name", "flopsy"),
					HaveKeyWithValue("instance_name", "csb-postgresql-mopsy"),
					HaveKeyWithValue("db_name", "cottontail"),
					HaveKeyWithValue("publicly_accessible", true),
					HaveKeyWithValue("region", "africa-north-4"),
					HaveKeyWithValue("storage_encrypted", true),
					HaveKeyWithValue("kms_key_id", "arn:aws:xxxx"),
					HaveKeyWithValue("multi_az", true),
					HaveKeyWithValue("allow_major_version_upgrade", false),
					HaveKeyWithValue("auto_minor_version_upgrade", false),
					HaveKeyWithValue("maintenance_window", "Mon:03:45-Mon:10:15"),
					HaveKeyWithValue("deletion_protection", true),
					HaveKeyWithValue("backup_retention_period", float64(2)),
					HaveKeyWithValue("backup_window", "01:02-03:04"),
					HaveKeyWithValue("copy_tags_to_snapshot", false),
					HaveKeyWithValue("delete_automated_backups", false),
					HaveKeyWithValue("monitoring_interval", float64(30)),
					HaveKeyWithValue("monitoring_role_arn", "arn:aws:iam::xxxxxxxxxxxx:role/enhanced_monitoring_access"),
					HaveKeyWithValue("performance_insights_enabled", true),
					HaveKeyWithValue("performance_insights_kms_key_id", "arn:aws:kms:us-west-2:649758297924:key/ebbb4ecc-ddfb-4e2f-8e93-c96d7bc43daa"),
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
				err := broker.Update(instanceID, serviceName, "small", params)

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update region", map[string]any{"region": "no-matter-what-region"}),
			Entry("update kms_key_id", map[string]any{"kms_key_id": "no-matter-what-key"}),
		)

		DescribeTable(
			"some allowed updates",
			func(key string, value any) {
				err := broker.Update(instanceID, serviceName, "small", map[string]any{key: value})

				Expect(err).NotTo(HaveOccurred())
			},
			Entry(nil, "require_ssl", true),
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
		)
	})
})
