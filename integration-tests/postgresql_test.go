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
		It("should check region constraints", func() {
			_, err := broker.Provision(serviceName, "small", map[string]any{"region": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(serviceName, "small", nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("cores", float64(2)),
					HaveKeyWithValue("engine_version", "11"),
					HaveKeyWithValue("storage_gb", float64(5)),
					HaveKeyWithValue("subsume", false),
					HaveKeyWithValue("require_ssl", true),
					HaveKeyWithValue("provider_verify_certificate", true),
					HaveKeyWithValue("storage_autoscale", false),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(0)),
					HaveKeyWithValue("parameter_group_name", ""),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-postgresql-%s", instanceID)),
					HaveKeyWithValue("db_name", "vsbdb"),
					HaveKeyWithValue("publicly_accessible", false),
					HaveKeyWithValue("region", "us-west-2"),
					HaveKeyWithValue("multi_az", false),
					HaveKeyWithValue("allow_major_version_upgrade", true),
					HaveKeyWithValue("auto_minor_version_upgrade", true),
					HaveKeyWithValue("maintenance_window", "Sun:00:00-Sun:00:00"),
					HaveKeyWithValue("deletion_protection", false),
					HaveKeyWithValue("backup_retention_period", float64(7)),
					HaveKeyWithValue("backup_window", "00:00-00:00"),
					HaveKeyWithValue("copy_tags_to_snapshot", true),
					HaveKeyWithValue("delete_automated_backups", true),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(serviceName, "small", map[string]any{
				"require_ssl":                 false,
				"provider_verify_certificate": false,
				"storage_autoscale":           true,
				"storage_autoscale_limit_gb":  float64(10),
				"parameter_group_name":        "flopsy",
				"instance_name":               "csb-postgresql-mopsy",
				"db_name":                     "cottontail",
				"publicly_accessible":         true,
				"region":                      "africa-north-4",
				"multi_az":                    true,
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
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("require_ssl", false),
					HaveKeyWithValue("provider_verify_certificate", false),
					HaveKeyWithValue("storage_autoscale", true),
					HaveKeyWithValue("storage_autoscale_limit_gb", float64(10)),
					HaveKeyWithValue("parameter_group_name", "flopsy"),
					HaveKeyWithValue("instance_name", "csb-postgresql-mopsy"),
					HaveKeyWithValue("db_name", "cottontail"),
					HaveKeyWithValue("publicly_accessible", true),
					HaveKeyWithValue("region", "africa-north-4"),
					HaveKeyWithValue("multi_az", true),
					HaveKeyWithValue("allow_major_version_upgrade", false),
					HaveKeyWithValue("auto_minor_version_upgrade", false),
					HaveKeyWithValue("maintenance_window", "Mon:03:45-Mon:10:15"),
					HaveKeyWithValue("deletion_protection", true),
					HaveKeyWithValue("backup_retention_period", float64(2)),
					HaveKeyWithValue("backup_window", "01:02-03:04"),
					HaveKeyWithValue("copy_tags_to_snapshot", false),
					HaveKeyWithValue("delete_automated_backups", false),
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

		It("should prevent updating region because it is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, serviceName, "small", map[string]any{"region": "no-matter-what-region"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))

			const initialProvisionInvocation = 1
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
		})

		DescribeTable(
			"some allowed updates",
			func(key string, value any) {
				err := broker.Update(instanceID, serviceName, "small", map[string]any{key: value})

				Expect(err).NotTo(HaveOccurred())
			},
			Entry(nil, "require_ssl", false),
			Entry(nil, "provider_verify_certificate", false),
			Entry(nil, "deletion_protection", true),
		)
	})
})
