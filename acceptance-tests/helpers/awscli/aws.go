// Package awscli provides test helpers for setting up AWS resources
package awscli

import (
	"fmt"
	"os/exec"
	"time"

	"code.cloudfoundry.org/jsonry"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func AWS(args ...string) []byte {
	cmd := exec.Command("aws", args...)
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "Running: %s\n", cmd.String())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(session).WithTimeout(time.Minute).Should(gexec.Exit(0))
	return session.Out.Contents()
}

func AWSToJSON[R any, pR *R](receiver pR, args ...string) {
	gomega.Expect(jsonry.Unmarshal(AWS(args...), receiver)).NotTo(gomega.HaveOccurred())
}
