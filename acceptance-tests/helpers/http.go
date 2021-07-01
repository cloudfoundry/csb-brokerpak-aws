package helpers

import (
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/gomega"
)

func HTTPGet(url string) string {
	response, err := http.Get(url)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusOK))

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	Expect(err).NotTo(HaveOccurred())

	return string(data)
}

func HTTPPost(url, data string) {
	response, err := http.Post(url, "text/html", strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated))
}
