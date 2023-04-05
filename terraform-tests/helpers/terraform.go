// Package helpers has helper functions for testing
package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	tfjson "github.com/hashicorp/terraform-json"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/gomega"
)

const (
	defaultTimeout = 35 * time.Minute
)

func Init(dir string) {
	command := exec.Command("terraform", "-chdir="+dir, "init")
	CommandStart(command)
}

func chdirFlag(dir string) string {
	return "-chdir=" + dir
}

func FailPlan(dir string, vars map[string]any) (*gexec.Session, error) {
	tfvarsPath := path.Join(dir, "terraform.tfvars.json")
	writeTFVarsFile(vars, tfvarsPath)
	defer os.Remove(tfvarsPath)

	tmpFile, err := os.CreateTemp(dir, "test-tf-plan")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	session, err := gexec.Start(terraformPlanCMD(dir, path.Base(tmpFile.Name())), GinkgoWriter, GinkgoWriter)
	if err != nil {
		return session, err
	}

	session = session.Wait(defaultTimeout)
	return session, nil
}

func ShowPlan(dir string, vars map[string]any) tfjson.Plan {
	tfvarsPath := path.Join(dir, "terraform.tfvars.json")
	writeTFVarsFile(vars, tfvarsPath)
	defer os.Remove(tfvarsPath)

	tmpFile, _ := os.CreateTemp(dir, "test-tf-plan")
	defer os.Remove(tmpFile.Name())
	CommandStart(terraformPlanCMD(dir, path.Base(tmpFile.Name())))

	jsonPlan := decodePlan(dir, path.Base(tmpFile.Name()))

	var plan tfjson.Plan
	err := json.Unmarshal(jsonPlan, &plan)
	Expect(err).NotTo(HaveOccurred())
	return plan
}

func terraformPlanCMD(dir string, planFile string) *exec.Cmd {
	return exec.Command("terraform", chdirFlag(dir), "plan", "-refresh=false", fmt.Sprintf("-out=%s", planFile))
}

func decodePlan(dir, planFile string) []byte {
	jsonPlan, err := CommandOutput(exec.Command("terraform", chdirFlag(dir), "show", "-json", planFile))
	Expect(err).ToNot(HaveOccurred())
	return jsonPlan
}

func CommandStart(command *exec.Cmd) *gexec.Session {
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, defaultTimeout).Should(gexec.Exit(0))
	return session
}

func writeTFVarsFile(vars map[string]any, tfvarsPath string) {
	variables, err := json.MarshalIndent(vars, "", "  ")
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(tfvarsPath, variables, 0755)
	Expect(err).ToNot(HaveOccurred())
}

func CommandOutput(command *exec.Cmd) ([]byte, error) {
	jsonOutput, err := command.Output()
	Expect(err).NotTo(HaveOccurred())
	return jsonOutput, err
}
