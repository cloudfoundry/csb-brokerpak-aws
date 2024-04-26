package main_test

import (
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTerraformProviderDynamodbns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TerraformProviderDynamodbns Suite")
}

func freePort() int {
	GinkgoHelper()

	listener, err := net.Listen("tcp", "localhost:0")
	Expect(err).NotTo(HaveOccurred())

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
