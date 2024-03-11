package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	sqsServiceID                  = "2198d694-bf85-11ee-a918-a7bdfa69a96d"
	sqsServiceName                = "csb-aws-sqs"
	sqsServiceDescription         = "Beta - CSB AWS SQS"
	sqsServiceDisplayName         = "CSB AWS SQS (Beta)"
	sqsServiceSupportURL          = "https://aws.amazon.com/sqs/"
	sqsServiceProviderDisplayName = "VMware"
	sqsCustomStandardPlanName     = "custom-standard"
	sqsCustomStandardPlanID       = "4c206ad6-bf89-11ee-8900-2f8e8940fc0d"
	sqsCustomFIFOPlanName         = "custom-fifo"
	sqsCustomFIFOPlanID           = "720feea2-c1bd-11ee-82ff-e3c9f193c356"
)

var customSQSPlans = []map[string]any{
	{
		"name":        sqsCustomStandardPlanName,
		"id":          sqsCustomStandardPlanID,
		"description": "Custom SQS standard queue plan",
		"metadata": map[string]any{
			"displayName": "custom-standard",
		},
	},
	{
		"name":        sqsCustomFIFOPlanName,
		"id":          sqsCustomFIFOPlanID,
		"description": "Custom SQS FIFO queue plan",
		"metadata": map[string]any{
			"displayName": "custom-fifo",
		},
	},
}

var _ = Describe("SQS", Label("SQS"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())

		DeferCleanup(func() {
			Expect(mockTerraform.Reset()).To(Succeed())
		})
	})

	It("should publish AWS SQS in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, sqsServiceName)
		Expect(service.ID).To(Equal(sqsServiceID))
		Expect(service.Description).To(Equal(sqsServiceDescription))
		Expect(service.Tags).To(ConsistOf("aws", "sqs", "beta"))
		Expect(service.Metadata.DisplayName).To(Equal(sqsServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(documentationURL))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.SupportUrl).To(Equal(sqsServiceSupportURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(sqsServiceProviderDisplayName))
		Expect(service.Plans).To(ConsistOf(
			MatchFields(IgnoreExtras, Fields{
				Name: Equal(sqsCustomStandardPlanName),
				ID:   Equal(sqsCustomStandardPlanID),
			}),
			MatchFields(IgnoreExtras, Fields{
				Name: Equal(sqsCustomFIFOPlanName),
				ID:   Equal(sqsCustomFIFOPlanID),
			}),
		))
	})

	Describe("provisioning", func() {
		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(sqsServiceName, sqsCustomStandardPlanName, params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
		)

		It("should provision a queue", func() {
			instanceID, err := broker.Provision(sqsServiceName, sqsCustomStandardPlanName, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{
						"pcf-instance-id": Equal(instanceID),
						"key1":            Equal("value1"),
						"key2":            Equal("value2"),
					})),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-sqs-%s", instanceID)),
					HaveKeyWithValue("fifo", BeFalse()),
					HaveKeyWithValue("visibility_timeout_seconds", BeNumerically("==", 30)),
					HaveKeyWithValue("message_retention_seconds", BeNumerically("==", 345600)),
					HaveKeyWithValue("max_message_size", BeNumerically("==", 262144)),
					HaveKeyWithValue("delay_seconds", BeZero()),
					HaveKeyWithValue("receive_wait_time_seconds", BeZero()),
					HaveKeyWithValue("region", fakeRegion),
					HaveKeyWithValue("aws_access_key_id", awsAccessKeyID),
					HaveKeyWithValue("aws_secret_access_key", awsSecretAccessKey),
					HaveKeyWithValue("dlq_arn", Equal("")),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(sqsServiceName, sqsCustomStandardPlanName, map[string]any{
				"region":                     "africa-north-4",
				"fifo":                       true,
				"visibility_timeout_seconds": 60,
				"message_retention_seconds":  60,
				"max_message_size":           1024,
				"delay_seconds":              600,
				"receive_wait_time_seconds":  20,
				"aws_access_key_id":          "fake-aws-access-key-id",
				"aws_secret_access_key":      "fake-aws-secret-access-key",
				"dlq_arn":                    "fake-arn",
				"max_receive_count":          5,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("region", "africa-north-4"),
					HaveKeyWithValue("fifo", BeTrue()),
					HaveKeyWithValue("visibility_timeout_seconds", BeNumerically("==", 60)),
					HaveKeyWithValue("message_retention_seconds", BeNumerically("==", 60)),
					HaveKeyWithValue("max_message_size", BeNumerically("==", 1024)),
					HaveKeyWithValue("delay_seconds", BeNumerically("==", 600)),
					HaveKeyWithValue("receive_wait_time_seconds", BeNumerically("==", 20)),
					HaveKeyWithValue("aws_access_key_id", "fake-aws-access-key-id"),
					HaveKeyWithValue("aws_secret_access_key", "fake-aws-secret-access-key"),
					HaveKeyWithValue("dlq_arn", "fake-arn"),
					HaveKeyWithValue("max_receive_count", BeNumerically("==", 5)),
				),
			)
		})

		It("should allow FIFO specific properties to be set on provision", func() {
			_, err := broker.Provision(sqsServiceName, sqsCustomFIFOPlanName, map[string]any{
				"fifo":                  true,
				"deduplication_scope":   "messageGroup",
				"fifo_throughput_limit": "perMessageGroupId",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("fifo", BeTrue()),
					HaveKeyWithValue("deduplication_scope", Equal("messageGroup")),
					HaveKeyWithValue("fifo_throughput_limit", Equal("perMessageGroupId")),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(sqsServiceName, sqsCustomStandardPlanName, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance",
			func(prop string, value any) {
				err := broker.Update(instanceID, sqsServiceName, sqsCustomStandardPlanName, map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update region", "region", "no-matter-what-region"),
			Entry("update fifo", "fifo", true),
			Entry("update dlq", "dlq", true),
		)

		DescribeTable(
			"some allowed updates",
			func(prop string, value any) {
				err := broker.Update(instanceID, sqsServiceName, sqsCustomStandardPlanName, map[string]any{prop: value})

				Expect(err).NotTo(HaveOccurred())
			},
			Entry(nil, "aws_access_key_id", "fake-aws-access-key-id"),
			Entry(nil, "aws_secret_access_key", "fake-aws-secret-access-key"),
			Entry(nil, "max_receive_count", 5),
			Entry(nil, "dlq_arn", "fake-arn"),
			Entry(nil, "visibility_timeout_seconds", 120),
			Entry(nil, "message_retention_seconds", 60),
			Entry(nil, "max_message_size", 1024),
			Entry(nil, "delay_seconds", 300),
			Entry(nil, "receive_wait_time_seconds", 15),
		)

		DescribeTable(
			"allowed FIFO updates",
			func(prop string, value any) {
				err := broker.Update(instanceID, sqsServiceName, sqsCustomFIFOPlanName, map[string]any{prop: value})

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update deduplication_scope", "deduplication_scope", "messageGroup"),
			Entry("update fifo_throughput_limit", "fifo_throughput_limit", "perMessageGroupId"),
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
					Value: "arn:aws:sqs::ap-northeast-3::example",
				},
				{
					Name:  "queue_name",
					Type:  "string",
					Value: "example_name",
				},
				{
					Name:  "queue_url",
					Type:  "string",
					Value: "example_url",
				},
				{
					Name:  "dlq",
					Type:  "boolean",
					Value: false,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			instanceID, err := broker.Provision(sqsServiceName, sqsCustomFIFOPlanName, nil)
			Expect(err).NotTo(HaveOccurred())

			bindResult, err := broker.Bind(sqsServiceName, sqsCustomFIFOPlanName, instanceID, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(
				Equal(map[string]any{
					"access_key_id":     "initial.access.key.id.test",
					"secret_access_key": "initial.secret.access.key.test",
					"region":            "ap-northeast-3",
					"arn":               "arn:aws:sqs::ap-northeast-3::example",
					"queue_name":        "example_name",
					"queue_url":         "example_url",
					"dlq":               false,
				}),
			)
		})
	})
})
