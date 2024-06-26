package main_test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/cloudfoundry/csb-brokerpak-aws/terraform-provider-dynamodbns/dynaclient"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
)

var _ = Describe("Resource dynamodbns_instance", func() {
	var session *gexec.Session
	var port int
	var localDynamoDBURL string
	var client *dynamodb.Client
	var prefix string

	BeforeEach(func() {
		var err error

		port = freePort()
		localDynamoDBURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		prefix = fmt.Sprintf("csb-%s-", uuid.New())

		pullCMD := exec.Command("docker", "pull", "amazon/dynamodb-local")
		sessionDockerPull, err := gexec.Start(pullCMD, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(sessionDockerPull).WithTimeout(time.Minute).WithPolling(time.Second).Should(gexec.Exit(0))

		cmd := exec.Command("docker", "run",
			"-p", fmt.Sprintf("%d:8000", port),
			"-t", "amazon/dynamodb-local")

		GinkgoWriter.Printf("running command: %s\n", cmd)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func(g Gomega) {
			client, err = dynaclient.New(context.TODO(), testRegion, "dummy", "dummy", localDynamoDBURL)
			g.Expect(err).NotTo(HaveOccurred())
			_, err = client.ListTables(context.TODO(), nil)
			g.Expect(err).NotTo(HaveOccurred())

		}).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(Succeed())

		populateTables(prefix, client)
	})

	AfterEach(func() {
		session.Terminate()
	})

	It("should clean up", func() {
		tfBody := fmt.Sprintf(`provider "csbdynamodbns" {
  region              = "%s"
  prefix              = "%s"
  custom_endpoint_url = "%s"
}

resource "csbdynamodbns_instance" "service_instance" {
  access_key_id     = "dummy"
  secret_access_key = "dummy"
}
		`, testRegion, prefix, localDynamoDBURL)
		applyHCL(tfBody, func(state *terraform.State) error {
			By("checking that only non-prefixed tables remain")
			Eventually(func(g Gomega) {
				tables, err := client.ListTables(context.TODO(), nil)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(tables.TableNames).To(HaveLen(3))
			}).WithTimeout(5 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
			return nil
		})
	})
})
