package terraformtests

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cp "github.com/otiai10/copy"
	"golang.org/x/exp/maps"
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

	var receiver struct {
		Region string `json:"region"`
	}
	Expect(json.Unmarshal([]byte(os.Getenv("GSB_PROVISION_DEFAULTS")), &receiver)).To(Succeed())
	if receiver.Region == "" {
		Fail("unable to determine region")
	}

	return receiver.Region
}

func getAWSConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecretAccessKey, ""))),
		config.WithRegion(awsRegion),
	)
	Expect(err).NotTo(HaveOccurred())
	return cfg
}

func pointer[A any](input A) *A {
	return &input
}

func safe[A any](input *A) (result A) {
	if input == nil {
		return
	}
	return *input
}
