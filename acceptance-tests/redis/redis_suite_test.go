package redis_test

import (
	"acceptancetests/helpers"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRedis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Redis Suite")
}

var serviceInstanceName string

var _ = BeforeSuite(func() {
	serviceInstanceName = helpers.RandomName("redis")
	helpers.CreateService("csb-aws-redis", "small", serviceInstanceName)
})

var _ = AfterSuite(func() {
	helpers.DeleteService(serviceInstanceName)
})
