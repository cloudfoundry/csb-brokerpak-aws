package helpers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func HTTPGet(url string) string {
	fmt.Fprintf(GinkgoWriter, "HTTP GET: %s\n", url)
	response, err := http.Get(url)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusOK))

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	Expect(err).NotTo(HaveOccurred())

	fmt.Fprintf(GinkgoWriter, "Recieved: %s\n", string(data))
	return string(data)
}

func HTTPPost(url, data string) {
	fmt.Fprintf(GinkgoWriter, "HTTP POST: %s\n", url)
	fmt.Fprintf(GinkgoWriter, "Sending data: %s\n", data)
	response, err := http.Post(url, "text/html", strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated))
}
