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
					"kms_master_key_id":                 BeNil(),
					"kms_data_key_reuse_period_seconds": BeNumerically("==", 300),
					"content_based_deduplication":       BeFalse(),
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

	Context("FIFO queues", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"fifo":                        true,
				"deduplication_scope":         "queue",
				"fifo_throughput_limit":       "perQueue",
				"content_based_deduplication": true,
			}))
		})

		It("should create an SQS FIFO queue with the correct properties", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                        Equal(fmt.Sprintf("%s.fifo", name)),
					"fifo_queue":                  BeTrue(),
					"deduplication_scope":         Equal("queue"),
					"fifo_throughput_limit":       Equal("perQueue"),
					"content_based_deduplication": BeTrue(),
				}),
			)
		})

		It("should create a FIFO SQS queue for high throughput mode", func() {
			customFIFOVars := map[string]any{
				"fifo":                        true,
				"deduplication_scope":         "messageGroup",
				"fifo_throughput_limit":       "perMessageGroupId",
				"content_based_deduplication": false,
			}
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, customFIFOVars))
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"fifo_queue":                  BeTrue(),
					"deduplication_scope":         Equal("messageGroup"),
					"fifo_throughput_limit":       Equal("perMessageGroupId"),
					"content_based_deduplication": BeFalse(),
				}),
			)
		})
	})

	Context("dead-letter queue", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"dlq_arn":           "arn:aws:sqs:us-west-2:123456789012:dlq",
				"max_receive_count": 3,
			}))
		})

		It("should create an SQS queue with the correct redrive policy", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"redrive_policy": Equal(`{"deadLetterTargetArn":"arn:aws:sqs:us-west-2:123456789012:dlq","maxReceiveCount":3}`),
				}),
			)
		})
	})

	Context("with visibility timeout set", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"visibility_timeout_seconds": 120,
			}))
		})

		It("should not be passed", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"visibility_timeout_seconds": BeNumerically("==", 120),
				}),
			)
		})
	})

	Context("with message retention set", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"message_retention_seconds": 1209600,
			}))
		})

		It("should not be passed", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"message_retention_seconds": BeNumerically("==", 1209600),
				}),
			)
		})
	})

	Context("with message size set", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"max_message_size": 1024,
			}))
		})

		It("should not be passed", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"max_message_size": BeNumerically("==", 1024),
				}),
			)
		})
	})

	Context("with delay set", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"delay_seconds": 300,
			}))
		})

		It("should not be passed", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"delay_seconds": BeNumerically("==", 300),
				}),
			)
		})
	})

	Context("with receive wait time set", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"receive_wait_time_seconds": 15,
			}))
		})

		It("should not be passed", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"receive_wait_time_seconds": BeNumerically("==", 15),
				}),
			)
		})
	})

	Context("with SQS-managed SSE disabled", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"sqs_managed_sse_enabled": false,
			}))
		})

		It("should disable SQS-managed server-side encryption", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"sqs_managed_sse_enabled": BeFalse(),
				}),
			)
		})
	})

	Context("with KMS master key specified", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"kms_master_key_id":                 "alias/aws/sqs",
				"kms_data_key_reuse_period_seconds": 300,
				"sqs_managed_sse_enabled":           true,
			}))
		})

		It("should use the specified KMS master key for encryption and data key reuse period specified", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"kms_master_key_id":                 Equal("alias/aws/sqs"),
					"kms_data_key_reuse_period_seconds": BeNumerically("==", 300),
				}),
			)
		})

		It("should override the value of `sqs_managed_sse_enabled`", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).NotTo(HaveKey("sqs_managed_sse_enabled"))
		})
	})
})
