package terraformtests

import (
	"os"
	"testing"

	"golang.org/x/exp/maps"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cp "github.com/otiai10/copy"
)

func TestTerraformTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TerraformTests Suite")
}

var (
	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsAccessKeyID     = os.Getenv("AWS_ACCESS_KEY_ID")
	awsVPCID           = "vpc-72464617"
	workingDir         = "../.aws-terraform-tests"
)

var _ = BeforeSuite(func() {
	err := os.MkdirAll(workingDir, os.ModePerm)
	Expect(err).ToNot(HaveOccurred())
	Expect(cp.Copy("../terraform", workingDir)).NotTo(HaveOccurred())
})

func buildVars(defaults, overrides map[string]any) map[string]any {
	result := map[string]any{}
	maps.Copy(result, defaults)
	maps.Copy(result, overrides)
	return result
}
