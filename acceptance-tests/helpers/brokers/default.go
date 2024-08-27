package brokers

import (
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"fmt"
	"os"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/gomega"
)

var defaultBrokerName string

func DefaultBrokerName() string {
	if defaultBrokerName != "" {
		return defaultBrokerName
	}

	var receiver struct {
		Names []string `jsonry:"resources.name"`
	}
	out, _ := cf.Run("curl", "/v3/service_brokers")
	Expect(jsonry.Unmarshal([]byte(out), &receiver)).To(Succeed())

	for _, brokerName := range receiver.Names {
		switch brokerName {
		case fmt.Sprintf("csb-%s", os.Getenv("USER")), "broker-cf-test", "cloud-service-broker-aws":
			defaultBrokerName = brokerName
			return brokerName
		}
	}

	panic("could not determine default broker name")
}
