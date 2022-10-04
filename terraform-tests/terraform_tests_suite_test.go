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
	workingDir string

	awsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsAccessKeyID     = os.Getenv("AWS_ACCESS_KEY_ID")
	awsVPCID           = "vpc-72464617"
)

var _ = BeforeSuite(func() {
	workingDir = GinkgoT().TempDir()
	Expect(cp.Copy("../terraform", workingDir)).NotTo(HaveOccurred())
})

func buildVars(defaults, overrides map[string]any) map[string]any {
	result := map[string]any{}
	maps.Copy(result, defaults)
	maps.Copy(result, overrides)
	return result
}
