package services

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"csbbrokerpakaws/acceptance-tests/helpers/cf"
)

func (s *ServiceInstance) Delete() {
	Delete(s.Name)
}

func Delete(name string) {
	session := cf.Start("delete-service", "-f", name, "--wait")
	Eventually(session, time.Hour).Should(Exit(0))
}
