package helpers

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func DeleteApps(names []string) {
	for i := 0; i < len(names); i++ {
		session := StartCF("delete", "-f", names[i])
		Eventually(session, time.Minute).Should(Exit(0))
	}
}
