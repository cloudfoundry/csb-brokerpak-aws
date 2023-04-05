package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTerraformProviderDynamodbns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TerraformProviderDynamodbns Suite")
}
