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
			"instance_name":              name,
			"fifo":                       false,
			"visibility_timeout_seconds": 30,
			"message_retention_seconds":  345600,
			"labels":                     map[string]string{"label1": "value1"},
			"aws_access_key_id":          awsAccessKeyID,
			"aws_secret_access_key":      awsSecretAccessKey,
			"region":                     awsRegion,
			"dlq_arn":                    "",
			"max_receive_count":          5,
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
					"name":                       Equal(name),
					"fifo_queue":                 BeFalse(),
					"visibility_timeout_seconds": BeNumerically("==", 30),
					"message_retention_seconds":  BeNumerically("==", 345600),
					"tags_all": MatchAllKeys(Keys{
						"label1": Equal("value1"),
					}),
				}),
			)
		})
	})

	Context("FIFO queues", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"fifo": true,
			}))
		})

		It("should create an SQS FIFO queue with the correct properties", func() {
			Expect(AfterValuesForType(plan, "aws_sqs_queue")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":       Equal(fmt.Sprintf("%s.fifo", name)),
					"fifo_queue": BeTrue(),
				}),
			)
		})
	})

	Context("with DLQ enabled", func() {
		BeforeAll(func() {
			dlqARN := "arn:aws:sqs:us-west-2:123456789012:dlq"
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"dlq_arn":           dlqARN,
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

	Context("with message retantion set", func() {
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
})
