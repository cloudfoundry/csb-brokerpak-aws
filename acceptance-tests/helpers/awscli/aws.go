// Package awscli provides test helpers for setting up AWS resources
package awscli

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"code.cloudfoundry.org/jsonry"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func AWS(args ...string) []byte {
	session := AWSSession(args...)
	gomega.Eventually(session).WithTimeout(time.Minute).Should(gexec.Exit(0))
	return session.Out.Contents()
}

// AWSQuery runs the AWS command with a JMESPath query, which extracts just the information
// that we need and results in much shorter logs when used to poll for a property value change
func AWSQuery(query string, args ...string) string {
	args = append(args, "--query", query)
	output := AWS(args...)
	return strings.Trim(strings.TrimSpace(string(output)), `"`)
}

func AWSToJSON[R any, pR *R](receiver pR, args ...string) {
	gomega.Expect(jsonry.Unmarshal(AWS(args...), receiver)).NotTo(gomega.HaveOccurred())
}

func AWSSession(args ...string) *gexec.Session {
	cmd := exec.Command("aws", args...)
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "Running: %s\n", cmd.String())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return session
}
