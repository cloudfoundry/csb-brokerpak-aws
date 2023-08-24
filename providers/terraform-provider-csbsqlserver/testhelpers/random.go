// Package testhelpers contains some helpers shared between different test packages
package testhelpers

import (
	"fmt"
	"net"
	"strings"

	"github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

func FreePort() int {
	listener, err := net.Listen("tcp", "localhost:0")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func RandomPassword() string {
	return randomWithPrefix("AaZz09~.")
}

func RandomDatabaseName() string {
	return randomWithPrefix("database")
}

func RandomTableName() string {
	return strings.ReplaceAll(randomWithPrefix("table"), "-", "_")
}

func RandomSchemaName() string {
	return strings.ReplaceAll(randomWithPrefix("schema"), "-", "_")
}

func randomWithPrefix(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New())
}
