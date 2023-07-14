package csbmajorengineversion_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTerraformProviderCSBMajorEngineVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Terraform Provider CSBMajorEngineVersion")
}
