package terraformtests

import (
	"path"

	. "csbbrokerpakaws/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("S3", Label("S3-terraform"), Ordered, func() {
	const bucketName = "csb-s3-test"

	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
		defaultVars           map[string]any
	)

	BeforeEach(func() {
		defaultVars = map[string]any{
			"aws_access_key_id":                          awsAccessKeyID,
			"aws_secret_access_key":                      awsSecretAccessKey,
			"bucket_name":                                bucketName,
			"region":                                     awsRegion,
			"acl":                                        "public-read",
			"enable_versioning":                          true,
			"boc_object_ownership":                       "BucketOwnerEnforced",
			"pab_block_public_acls":                      false,
			"pab_block_public_policy":                    false,
			"pab_ignore_public_acls":                     false,
			"pab_restrict_public_buckets":                false,
			"sse_default_kms_key_id":                     nil,
			"sse_extra_kms_key_ids":                      nil,
			"sse_default_algorithm":                      nil,
			"sse_bucket_key_enabled":                     false,
			"ol_enabled":                                 false,
			"ol_configuration_default_retention_enabled": nil,
			"ol_configuration_default_retention_mode":    nil,
			"ol_configuration_default_retention_days":    nil,
			"ol_configuration_default_retention_years":   nil,
			"labels":      map[string]any{"k1": "v1"},
			"require_tls": false,
		}
	})

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "s3/provision")
		Init(terraformProvisionDir)
	})

	Context("with default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(5))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_s3_bucket",
				"aws_s3_bucket_acl",
				"aws_s3_bucket_versioning",
				"aws_s3_bucket_ownership_controls",
				"aws_s3_bucket_public_access_block",
			))
		})

		It("should create an S3 bucket with the correct properties", func() {
			Expect(AfterValuesForType(plan, "aws_s3_bucket")).To(
				MatchKeys(IgnoreExtras, Keys{
					"bucket":              Equal(bucketName),
					"object_lock_enabled": BeFalse(),
					"tags": MatchAllKeys(Keys{
						"k1": Equal("v1"),
					}),
				}),
			)
		})

		It("should create an S3 bucket ACL resource with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_s3_bucket_acl")).To(
				MatchKeys(IgnoreExtras, Keys{
					"acl": Equal("public-read"),
				}),
			)
		})

		It("should create an S3 bucket versioning with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_s3_bucket_versioning")).To(
				MatchKeys(IgnoreExtras, Keys{
					"versioning_configuration": ConsistOf(
						MatchKeys(IgnoreExtras, Keys{
							"status": Equal("Enabled"),
						}),
					),
				}),
			)
		})

		It("should create an S3 bucket ownership controls resource with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_s3_bucket_ownership_controls")).To(
				MatchKeys(IgnoreExtras, Keys{
					"rule": ConsistOf(
						MatchKeys(IgnoreExtras, Keys{
							"object_ownership": Equal("BucketOwnerEnforced"),
						}),
					),
				}))
		})

		It("should create an S3 bucket public access block resource with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_s3_bucket_public_access_block")).To(
				MatchKeys(IgnoreExtras, Keys{
					"block_public_acls":       BeFalse(),
					"block_public_policy":     BeFalse(),
					"ignore_public_acls":      BeFalse(),
					"restrict_public_buckets": BeFalse(),
				}))
		})
	})

	Context("setting require_tls to true", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{"require_tls": true}))
		})

		It("should create an aws_s3_bucket_policy", func() {
			Expect(plan.ResourceChanges).To(HaveLen(6))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_s3_bucket",
				"aws_s3_bucket_acl",
				"aws_s3_bucket_versioning",
				"aws_s3_bucket_ownership_controls",
				"aws_s3_bucket_public_access_block",
				"aws_s3_bucket_policy",
			))
		})
	})
})
