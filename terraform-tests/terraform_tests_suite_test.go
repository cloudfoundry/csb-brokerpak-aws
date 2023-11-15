package terraformtests

import (
	"os"
	"regexp"
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
	awsVPCID           = os.Getenv("AWS_PAS_VPC_ID")
	awsRegion          = getAWSRegion()
)

var _ = BeforeSuite(func() {
	workingDir = GinkgoT().TempDir()
	Expect(cp.Copy("../terraform", workingDir)).NotTo(HaveOccurred())
})

func buildVars(varOverrides ...map[string]any) map[string]any {
	result := map[string]any{}
	for _, override := range varOverrides {
		maps.Copy(result, override)
	}
	return result
}

func getAWSRegion() string {
	envRegion := os.Getenv("AWS_DEFAULT_REGION")
	if envRegion != "" {
		return envRegion
	}

	return getAWSRegionFromCSBDefaults()
}

func getAWSRegionFromCSBDefaults() string {
	r := regexp.MustCompile(`"region":\s*"([a-z0-9-]+)"`)
	matches := r.FindStringSubmatch(os.Getenv("GSB_PROVISION_DEFAULTS"))
	if matches != nil {
		return matches[1]
	}
	return ""
}
