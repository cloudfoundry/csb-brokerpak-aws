// Package testhelpers contains some helpers shared between different test packages
package testhelpers

import (
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/onsi/gomega"
)

func FreePort() int {
	listener, err := net.Listen("tcp", "localhost:0")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func RandomPassword() string {
	return fmt.Sprintf("AaZz09~._%s", uuid.New())
}

func RandomDatabaseName() string {
	return fmt.Sprintf("database_%s", uuid.New())
}
