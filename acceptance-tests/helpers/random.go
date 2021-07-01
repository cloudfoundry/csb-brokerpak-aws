package helpers

import (
	"crypto/rand"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/Pallinder/go-randomdata"
)

func RandomName(prefixes ...string) string {
	return strings.Join(append(prefixes, randomdata.Adjective(), randomdata.Noun()), "-")
}

func RandomString() string {
	const length = 10
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	Expect(err).NotTo(HaveOccurred())
	return fmt.Sprintf("%x", buf)
}
