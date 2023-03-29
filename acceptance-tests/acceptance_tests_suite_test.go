package acceptance_tests_test

import (
	"csbbrokerpakaws/acceptance-tests/helpers/environment"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAcceptanceTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AcceptanceTests Suite")
}

var metadata environment.Metadata

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
})

func must[A any](input A, err error) A {
	GinkgoHelper()

	if err != nil {
		Fail(fmt.Sprintf("unexpected error: %s", err.Error()))
	}

	return input
}
