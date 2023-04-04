package csbdynamodbns_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCsbdynamodbns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSB DynamoDB-NS Suite")
}
