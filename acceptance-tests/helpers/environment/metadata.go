// Package environment manages environment variables
package environment

import (
	"encoding/json"
	"os"

	"github.com/onsi/gomega"
)

type Metadata struct {
	Name   string `json:"name"`
	VPC    string `json:"pas_vpc_id"`
	Region string `json:"region"`
}

func ReadMetadata() Metadata {
	file := os.Getenv("ENVIRONMENT_LOCK_METADATA")
	gomega.Expect(file).NotTo(gomega.BeEmpty(), "You must set the ENVIRONMENT_LOCK_METADATA environment variable")

	contents, err := os.ReadFile(file)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	var metadata Metadata
	gomega.Expect(json.Unmarshal(contents, &metadata)).NotTo(gomega.HaveOccurred())
	gomega.Expect(metadata.Name).NotTo(gomega.BeEmpty())
	gomega.Expect(metadata.VPC).NotTo(gomega.BeEmpty())
	gomega.Expect(metadata.Region).NotTo(gomega.BeEmpty())
	return metadata
}
