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
	"storage_gb":    10,
	"subsume":       false,
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
		It("should check region constraints", func() {
			_, err := broker.Provision(serviceName, "small", map[string]any{"region": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
		})

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(serviceName, customMySQLPlan["name"].(string), nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("use_tls", true),
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
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(serviceName, "custom-sample", map[string]any{
				"use_tls":                     false,
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
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("engine", "mysql"),
					HaveKeyWithValue("engine_version", "8"),
					HaveKeyWithValue("cores", float64(4)),
					HaveKeyWithValue("storage_gb", float64(10)),
					HaveKeyWithValue("use_tls", false),
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
	})
})
