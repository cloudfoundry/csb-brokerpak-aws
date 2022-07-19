package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var customS3Plans = []map[string]any{
	customS3Plan,
	customS3PlanWithACL,
}

var customS3Plan = map[string]any{
	"name":        "custom-plan",
	"id":          "9dfa265e-1c4d-40c6-ade6-b341ffd6ccc3",
	"description": "Beta - custom S3 plan defined by customer",
	"metadata": map[string]any{
		"displayName": "custom S3 service (Beta)",
	},
}

var customS3PlanWithACL = map[string]any{
	"name":        "custom-plan-with-acl",
	"acl": "private",
	"id":          "9dfa265e-1c4d-40c6-ade6-b341ffd6ccc4",
	"description": "Beta - custom S3 plan defined by customer specifying acl",
	"metadata": map[string]any{
		"displayName": "custom S3 service with acl (Beta)",
	},
}

var _ = Describe("S3", Label("s3"), func() {
	const s3ServiceName = "csb-aws-s3-bucket"
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish AWS S3 in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, s3ServiceName)
		Expect(service.ID).NotTo(BeNil())
		Expect(service.Name).NotTo(BeNil())
		Expect(service.Tags).To(ConsistOf("aws", "s3", "beta"))
		Expect(service.Metadata.ImageUrl).NotTo(BeNil())
		Expect(service.Metadata.DisplayName).NotTo(BeNil())
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{"Name": Equal("private")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("public-read")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("custom-plan")}),
				MatchFields(IgnoreExtras, Fields{"Name": Equal("custom-plan-with-acl")}),
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
		It("should provision private plan", func() {
			instanceID, err := broker.Provision(s3ServiceName, "private", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("bucket_name", "csb-"+instanceID),
					HaveKeyWithValue("enable_versioning", false),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
					HaveKeyWithValue("region", "us-west-2"),
					HaveKeyWithValue("acl", "private"),
					HaveKeyWithValue("aws_access_key_id", awsAccessKeyID),
					HaveKeyWithValue("aws_secret_access_key", awsSecretAccessKey),
				),
			)
		})

		It("should allow setting properties not defined in the plan", func() {
			instanceID, err := broker.Provision(s3ServiceName, customS3Plan["name"].(string), map[string]any{
				"bucket_name":           "fake-bucket-name",
				"enable_versioning":     true,
				"region":                "eu-west-1",
				"acl":                "public-read",
				"aws_access_key_id":     "fake-aws-access-key-id",
				"aws_secret_access_key": "fake-aws-secret-access-key",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("bucket_name", "fake-bucket-name"),
					HaveKeyWithValue("enable_versioning", true),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
					HaveKeyWithValue("region", "eu-west-1"),
					HaveKeyWithValue("acl", "public-read"),
					HaveKeyWithValue("aws_access_key_id", "fake-aws-access-key-id"),
					HaveKeyWithValue("aws_secret_access_key", "fake-aws-secret-access-key"),
				),
			)
		})

		It("should not allow changing of plan defined properties", func() {
			_, err := broker.Provision(s3ServiceName, "private", map[string]any{"acl": "public-read"})

			Expect(err).To(MatchError(ContainSubstring("plan defined properties cannot be changed: acl")))
		})

		Describe("property constraints", func() {
			It("should validate invalid characters in the region parameter", func() {
				_, err := broker.Provision(s3ServiceName, customS3Plan["name"].(string), map[string]any{"region": "-Asia-northeast1"})

				Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
			})
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(s3ServiceName, customS3Plan["name"].(string), map[string]any{"acl": "private"})

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should allow updating properties not flagged as `prohibit_update` and not specified in the plan",
			func(params map[string]any) {
				err := broker.Update(instanceID, s3ServiceName, customS3Plan["name"].(string), params)

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update aws_access_key_id", map[string]any{"aws_access_key_id": "another-aws_access_key_id"}),
			Entry("update aws_secret_access_key", map[string]any{"aws_secret_access_key": "another-aws_secret_access_key"}),
			Entry("update acl", map[string]any{"acl": "public-read"}),
		)

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance and lost data",
			func(params map[string]any) {
				err := broker.Update(instanceID, s3ServiceName, customS3Plan["name"].(string), params)

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update enable_versioning", map[string]any{"enable_versioning": false}),
			Entry("update region", map[string]any{"region": "no-matter-what-region"}),
			Entry("update bucket name", map[string]any{"bucket_name": "some-nicer-name"}),
		)

		It("should not allow updating properties that are specified in the plan", func() {
			err := broker.Update(instanceID, s3ServiceName, customS3PlanWithACL["name"].(string), map[string]any{"acl": "public-read"})

			Expect(err).To(
				MatchError(
					ContainSubstring("plan defined properties cannot be changed: acl"),
				),
			)
		})

		DescribeTable("should not allow updating additional properties",
			func(key string, value any) {
				err := broker.Update(instanceID, s3ServiceName, customS3Plan["name"].(string), map[string]any{key: value})

				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("additional properties are not allowed: %s", key),
						),
					),
				)
			},
			Entry("update name", "name", "fake-name"),
			Entry("update id", "id", "fake-id"),
		)
	})

	Describe("bind a service ", func() {
		It("return the bind values from terraform output", func() {
			err := mockTerraform.SetTFState([]testframework.TFStateValue{
				{
					Name:  "access_key_id",
					Type:  "string",
					Value: "initial.access.key.id.test",
				},
				{
					Name:  "secret_access_key",
					Type:  "string",
					Value: "initial.secret.access.key.test",
				},
				{
					Name:  "region",
					Type:  "string",
					Value: "ap-northeast-3",
				},
				{
					Name:  "arn",
					Type:  "string",
					Value: "arn:aws:s3:::examplebucket/developers/design_info.doc",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			instanceID, err := broker.Provision(s3ServiceName, customS3Plan["name"].(string), nil)
			Expect(err).NotTo(HaveOccurred())

			err = mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "access_key_id", Type: "string", Value: "subsequent.access.key.id.test"},
				{Name: "secret_access_key", Type: "string", Value: "subsequent.secret.access.key.test"},
			})
			Expect(err).NotTo(HaveOccurred())

			bindResult, err := broker.Bind(s3ServiceName, customS3Plan["name"].(string), instanceID, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(
				Equal(map[string]any{
					"access_key_id":     "subsequent.access.key.id.test",
					"secret_access_key": "subsequent.secret.access.key.test",
					"region":            "ap-northeast-3",
					"arn":               "arn:aws:s3:::examplebucket/developers/design_info.doc",
				}),
			)
		})
	})
})
