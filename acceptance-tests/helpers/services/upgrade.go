package services

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"csbbrokerpakaws/acceptance-tests/helpers/cf"
)

func (s *ServiceInstance) Upgrade() {
	if !s.UpgradeAvailable() {
		fmt.Printf("no upgrade available for service instance\n")
		return
	}

	var command []string
	switch cf.Version() {
	case cf.VersionV8:
		command = []string{"upgrade-service", s.Name, "--force"}
	default:
		command = []string{"update-service", s.Name, "--upgrade", "--force"}
	}

	session := cf.Start(command...)
	Eventually(session).WithTimeout(asyncCommandTimeout).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", s.Name)
		Expect(out).NotTo(MatchRegexp(`status:\s+update failed`))
		return out
	}).WithTimeout(operationTimeout).WithPolling(pollingInterval).Should(MatchRegexp(`status:\s+update succeeded`))

	Expect(s.UpgradeAvailable()).To(BeFalse(), "service instance has an upgrade available after upgrade")
}

func (s *ServiceInstance) UpgradeAvailable() bool {
	out, _ := cf.Run("curl", fmt.Sprintf("/v3/service_instances/%s", s.GUID()))

	var receiver struct {
		UpgradeAvailable bool `json:"upgrade_available"`
	}
	Expect(json.Unmarshal([]byte(out), &receiver)).NotTo(HaveOccurred())
	return receiver.UpgradeAvailable
}
