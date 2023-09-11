// Package jdbcapp contains data structures used for decoding
// responses from the JDBC test application, and relevant test helpers
package jdbcapp

import (
	"fmt"
	"os"
	"path"

	. "github.com/onsi/gomega"
)

type DatabaseEngine string

const (
	MySQL      DatabaseEngine = "mysql"
	PostgreSQL DatabaseEngine = "postgres"
	SQLServer  DatabaseEngine = "sqlserver"
)

type AppResponseUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PostgresSSLInfo struct {
	Pid          int    `json:"pid"`
	SSL          bool   `json:"ssl"`
	Version      string `json:"version"`
	Cipher       string `json:"cipher"`
	Bits         int    `json:"bits"`
	ClientDN     string `json:"clientDN"`
	ClientSerial string `json:"clientSerial"`
	IssuerDN     string `json:"issuerDN"`
}

type MySQLSSLInfo struct {
	VariableName string `json:"variableName"`
	Value        string `json:"value"`
}

func ManifestFor(dbEngine DatabaseEngine) string {
	testExecutable, err := os.Executable()
	Expect(err).NotTo(HaveOccurred())

	testPath := path.Dir(testExecutable)
	return path.Join(testPath, "apps", "jdbctestapp", fmt.Sprintf("manifest-%s.yml", dbEngine))

}
