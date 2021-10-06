package helpers

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

type EnvVar struct {
	Name  string
	Value interface{}
}

func SetBrokerEnv(brokerName string, envVars ...EnvVar) {
	for _, envVar := range envVars {
		switch v := envVar.Value.(type) {
		case string:
			if v == "" {
				CF("unset-env", brokerName, envVar.Name)
			} else {
				CF("set-env", brokerName, envVar.Name, v)
			}
		default:
			data, err := json.Marshal(v)
			Expect(err).NotTo(HaveOccurred())
			CF("set-env", brokerName, envVar.Name, string(data))
		}
	}
}

func SetBrokerEnvAndRestart(envVars ...EnvVar) {
	const broker = "cloud-service-broker-aws"

	SetBrokerEnv(broker, envVars...)
	session := StartCF("restart", broker)
	waitForAppPush(session, broker)
}
