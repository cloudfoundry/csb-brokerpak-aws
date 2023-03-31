package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// GET does an HTTP get, returning the body as a payload
func (a *App) GET(format string, s ...any) Payload {
	GinkgoHelper()

	response := a.GETResponse(format, s...)
	Expect(response).To(HaveHTTPStatus(http.StatusOK))

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	payload := NewPayload(response)
	GinkgoWriter.Printf("Received: %s\n", payload.String())
	return payload
}

// GETResponse does an HTTP get, returning the *http.Response
func (a *App) GETResponse(format string, s ...any) *http.Response {
	GinkgoHelper()

	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP GET: %s\n", url)
	response, err := http.Get(url)
	Expect(err).NotTo(HaveOccurred())
	return response
}

func (a *App) PUT(input any, format string, s ...any) {
	GinkgoHelper()

	url := a.urlf(format, s...)
	data := stringify(input)
	GinkgoWriter.Printf("HTTP PUT: %s\n", url)
	GinkgoWriter.Printf("Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Set("Content-Type", "text/html")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated, http.StatusOK))
}

func (a *App) POSTResponse(data, format string, s ...any) *http.Response {
	GinkgoHelper()

	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP POST: %s\n", url)
	GinkgoWriter.Printf("Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	return response
}

func (a *App) POST(input any, format string, s ...any) Payload {
	GinkgoHelper()

	url := a.urlf(format, s...)
	data := stringify(input)
	GinkgoWriter.Printf("HTTP POST: %s\n", url)
	GinkgoWriter.Printf("Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated, http.StatusOK, http.StatusAccepted))

	defer response.Body.Close()
	return NewPayload(response)
}

func (a *App) DELETEResponse(format string, s ...any) (*http.Response, error) {
	GinkgoHelper()

	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP DELETE: %s\n", url)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	Expect(err).NotTo(HaveOccurred())
	return http.DefaultClient.Do(request)
}

func (a *App) DELETE(format string, s ...any) {
	GinkgoHelper()

	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP DELETE: %s\n", url)
	request, err2 := http.NewRequest(http.MethodDelete, url, nil)
	Expect(err2).NotTo(HaveOccurred())
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusGone, http.StatusNoContent, http.StatusOK, http.StatusNotFound))
}

func (a *App) urlf(format string, s ...any) string {
	GinkgoHelper()

	base := a.URL
	path := fmt.Sprintf(format, s...)
	switch {
	case len(path) == 0:
		return base
	case path[0] != '/':
		return fmt.Sprintf("%s/%s", base, path)
	default:
		return base + path
	}
}

func stringify(input any) string {
	switch i := input.(type) {
	case string:
		return i
	case []byte:
		return string(i)
	case nil:
		return ""
	}

	data, err := json.Marshal(input)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}
