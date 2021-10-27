package dynamodb_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDynamoDB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DynamoDB Suite")
}
