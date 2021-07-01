package helpers

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func PushAppsUnstarted(prefix, appDir string, copies int) []string {
	names := make([]string, copies)

	for i := 0; i < copies; i++ {
		names[i] = RandomName(prefix)
		session := StartCF("push", "--no-start", "-b", "binary_buildpack", "-p", appDir, names[i])
		Eventually(session, 5*time.Minute).Should(Exit(0))
	}

	return names
}
