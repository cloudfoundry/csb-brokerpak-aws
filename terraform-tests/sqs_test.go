package terraformtests

import (
	. "csbbrokerpakaws/terraform-tests/helpers"
	"fmt"
	"path"
	"time"

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
			"instance_name":         name,
			"fifo":                  false,
			"labels":                map[string]string{"label1": "value1"},
			"aws_access_key_id":     awsAccessKeyID,
			"aws_secret_access_key": awsSecretAccessKey,
			"region":                awsRegion,
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
					"name":       Equal(name),
					"fifo_queue": BeFalse(),
					"tags": MatchAllKeys(Keys{
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
})
