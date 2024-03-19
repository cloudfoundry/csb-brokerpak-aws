package terraformtests

import (
	"fmt"
	"path"
	"time"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("SQS", Label("SQS-terraform"), Ordered, func() {
	var (
		name                  string
		plan                  tfjson.Plan
		terraformProvisionDir string
		defaultVars           map[string]any
	)

	BeforeAll(func() {
		name = fmt.Sprintf("csb-tf-test-sqs-%d-%d", GinkgoRandomSeed(), time.Now().Unix())

		terraformProvisionDir = path.Join(workingDir, "sqs/provision")
		Init(terraformProvisionDir)
	})

	BeforeEach(func() {
		defaultVars = map[string]any{
			"instance_name":                     name,
			"fifo":                              false,
			"visibility_timeout_seconds":        30,
			"message_retention_seconds":         345600,
			"max_message_size":                  262144,
			"delay_seconds":                     0,
			"receive_wait_time_seconds":         0,
			"labels":                            map[string]string{"label1": "value1"},
			"aws_access_key_id":                 awsAccessKeyID,
			"aws_secret_access_key":             awsSecretAccessKey,
			"region":                            awsRegion,
			"dlq_arn":                           "",
			"max_receive_count":                 5,
			"deduplication_scope":               nil,
			"fifo_throughput_limit":             nil,
			"content_based_deduplication":       false,
			"sqs_managed_sse_enabled":           true,
			"kms_master_key_id":                 "",
			"kms_extra_key_ids":                 "",
			"kms_data_key_reuse_period_seconds": 300,
		}
	})

	Context("with default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(1))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_sqs_queue",
			))
		})

		It("should create an SQS queue with the correct properties", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                              Equal(name),
					"fifo_queue":                        BeFalse(),
					"visibility_timeout_seconds":        BeNumerically("==", 30),
					"message_retention_seconds":         BeNumerically("==", 345600),
					"max_message_size":                  BeNumerically("==", 262144),
					"delay_seconds":                     BeZero(),
					"receive_wait_time_seconds":         BeZero(),
					"content_based_deduplication":       BeFalse(),
					"kms_master_key_id":                 BeNil(),
					"kms_data_key_reuse_period_seconds": BeNumerically("==", 300),
					"sqs_managed_sse_enabled":           BeTrue(),
					"tags_all": MatchAllKeys(Keys{
						"label1": Equal("value1"),
					}),
				}),
			)

			Expect(AfterValuesForType(plan, "aws_sqs_queue")).NotTo(SatisfyAny(
				HaveKey("redrive_policy"),
				HaveKey("deduplication_scope"),
				HaveKey("fifo_throughput_limit"),
			))
		})
	})

	Context("with non-default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"fifo":                              true,
				"visibility_timeout_seconds":        31,
				"message_retention_seconds":         300000,
				"max_message_size":                  200000,
				"delay_seconds":                     1,
				"receive_wait_time_seconds":         2,
				"dlq_arn":                           "fake-dlq-arn",
				"max_receive_count":                 4,
				"deduplication_scope":               "messageGroup",
				"fifo_throughput_limit":             "perMessageGroupId",
				"content_based_deduplication":       true,
				"sqs_managed_sse_enabled":           false,
				"kms_master_key_id":                 "fake-key-id",
				"kms_extra_key_ids":                 "fake-extra-key-id-1,fake-extra-key-id-2",
				"kms_data_key_reuse_period_seconds": 231,
			}))
		})

		It("should reflect the non-default values", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"fifo_queue":                        BeTrue(),
					"visibility_timeout_seconds":        BeNumerically("==", 31),
					"message_retention_seconds":         BeNumerically("==", 300000),
					"max_message_size":                  BeNumerically("==", 200000),
					"delay_seconds":                     BeNumerically("==", 1),
					"receive_wait_time_seconds":         BeNumerically("==", 2),
					"redrive_policy":                    MatchJSON(`{"deadLetterTargetArn":"fake-dlq-arn","maxReceiveCount":4}`),
					"content_based_deduplication":       BeTrue(),
					"deduplication_scope":               Equal("messageGroup"),
					"fifo_throughput_limit":             Equal("perMessageGroupId"),
					"kms_master_key_id":                 Equal("fake-key-id"),
					"kms_data_key_reuse_period_seconds": BeNumerically("==", 231),
				}),
			)

			Expect(AfterValuesForType(plan, "aws_sqs_queue")).NotTo(HaveKey("sqs_managed_sse_enabled"))
		})
	})
})
