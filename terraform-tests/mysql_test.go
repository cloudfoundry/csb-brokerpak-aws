package terraformtests

import (
	. "csbbrokerpakaws/terraform-tests/helpers"
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("mysql", Label("mysql-terraform"), func() {

	defaultVars := map[string]any{
		"cores":                       nil,
		"instance_name":               "csb-mysql-test",
		"db_name":                     "vsbdb",
		"labels":                      map[string]string{"label1": "value1"},
		"storage_gb":                  5,
		"storage_type":                "io1",
		"iops":                        3000,
		"publicly_accessible":         false,
		"multi_az":                    false,
		"instance_class":              "an-instance-class",
		"engine":                      "mysql",
		"engine_version":              5.7,
		"aws_vpc_id":                  awsVPCID,
		"storage_autoscale":           false,
		"storage_autoscale_limit_gb":  0,
		"storage_encrypted":           false,
		"parameter_group_name":        "",
		"rds_subnet_group":            "",
		"rds_vpc_security_group_ids":  "",
		"allow_major_version_upgrade": true,
		"auto_minor_version_upgrade":  true,
		"maintenance_end_hour":        nil,
		"maintenance_start_hour":      nil,
		"maintenance_end_min":         nil,
		"maintenance_start_min":       nil,
		"maintenance_day":             nil,
		"use_tls":                     true,
		"deletion_protection":         false,
		"backup_retention_period":     7,
		"backup_window":               nil,
		"copy_tags_to_snapshot":       true,
		"delete_automated_backups":    true,
		"aws_access_key_id":           awsAccessKeyID,
		"aws_secret_access_key":       awsSecretAccessKey,
		"region":                      "us-west-2",
		"subsume":                     false,
	}

	Describe("mysql provision", func() {
		var terraformProvisionDir string
		var plan tfjson.Plan
		BeforeEach(OncePerOrdered, func() {
			terraformProvisionDir = path.Join(workingDir, "mysql/provision")
			Init(terraformProvisionDir)
		})

		Context("provisions an instance", Ordered, func() {
			Context("mysql parameter groups", func() {

				Context("No parameter group name passed", func() {
					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
					})

					It("should use the default parameter group", func() {
						Expect(ResourceCreationForType(plan, "aws_db_parameter_group")).To(BeEmpty())

						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"parameter_group_name": Equal("default.mysql5.7"),
							}))
					})

				})

				Context("Parameter group passed", func() {

					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
							"parameter_group_name": "some-parameter-group-name",
						}))
					})

					It("should not create a parameter group if name is provided", func() {
						Expect(ResourceCreationForType(plan, "aws_db_parameter_group")).To(BeEmpty())

						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"parameter_group_name": Equal("some-parameter-group-name"),
							}))
					})
				})

			})

			Context("storage type", func() {
				Context("default values", func() {
					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
					})

					It("default values work with io1 and 3000 iops", func() {
						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"storage_type": Equal("io1"),
								"iops":         Equal(float64(3000)),
							}))
					})
				})

				Context("storage_type gp2", func() {
					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
							"storage_type": "gp2",
						}))
					})

					It("iops should be null", func() {
						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"storage_type": Equal("gp2"),
								"iops":         BeNil(),
							}))
					})
				})

			})

			Context("autoscaling", func() {
				Context("storage_autoscale is false", func() {
					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
							"storage_autoscale":          false,
							"storage_autoscale_limit_gb": 200,
						}))
					})

					It("autoscaling should be disabled", func() {
						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"max_allocated_storage": BeNil(),
							}))
					})
				})

				Context("storage_autoscale is true and limit > storage_gb", func() {
					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
							"storage_autoscale":          true,
							"storage_autoscale_limit_gb": 200,
						}))
					})

					It("autoscaling should be enabled", func() {
						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"max_allocated_storage": Equal(float64(200)),
							}))
					})
				})

				Context("storage_autoscale is true and limit <= storage_gb", func() {
					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
							"storage_autoscale":          true,
							"storage_autoscale_limit_gb": 5,
						}))
					})

					It("autoscaling should be disabled", func() {
						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"max_allocated_storage": BeNil(),
							}))
					})
				})

			})

			Context("security groups", func() {
				Context("no security group ids passed", func() {
					It("should create a new one", func() {
						Expect(UnknownValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"vpc_security_group_ids": BeTrue(),
							}))
						Expect(ResourceCreationForType(plan, "aws_security_group")).To(HaveLen(1))
						Expect(AfterValuesForType(plan, "aws_security_group")).To(
							MatchKeys(IgnoreExtras, Keys{
								"name": Equal("csb-mysql-test-sg"),
							}))
					})
				})

				Context("security group ids passed", func() {
					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
							"rds_vpc_security_group_ids": "group1,group2,group3",
						}))
					})

					It("should use the ids passed and no create new security groups", func() {
						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"vpc_security_group_ids": ConsistOf("group1", "group2", "group3"),
							}))
						Expect(ResourceCreationForType(plan, "aws_security_group")).To(BeEmpty())
					})
				})
			})

			Context("maintenance_window", func() {
				BeforeEach(func() {
					plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
				})
				Context("no window", func() {
					It("should not be passed", func() {
						Expect(AfterValuesForType(plan, "aws_db_instance")).To(Not(HaveKey("maintenance_window")))
					})
				})

				Context("only maintenance_day specified", func() {
					BeforeEach(func() {
						plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
							"maintenance_day":        "Mon",
							"maintenance_start_hour": "01",
							"maintenance_end_hour":   "02",
							"maintenance_start_min":  "03",
							"maintenance_end_min":    "04",
						}))
					})

					It("should pass the correct window", func() {
						Expect(AfterValuesForType(plan, "aws_db_instance")).To(
							MatchKeys(IgnoreExtras, Keys{
								"maintenance_window": Equal("mon:01:03-mon:02:04"),
							}))
					})

				})

			})

		})
	})
})
