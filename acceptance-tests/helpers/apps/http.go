package apps

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func (a *App) GET(format string, s ...any) string {
	response, err := a.GETResponse(format, s...)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusOK))

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	data, err := io.ReadAll(response.Body)
	Expect(err).NotTo(HaveOccurred())

	_, _ = fmt.Fprintf(GinkgoWriter, "Recieved: %s\n", string(data))
	return string(data)
}

func (a *App) GETResponse(format string, s ...any) (*http.Response, error) {
	url := a.urlf(format, s...)
	_, _ = fmt.Fprintf(GinkgoWriter, "HTTP GET: %s\n", url)
	return http.Get(url)
}

func (a *App) PUT(data, format string, s ...any) {
	url := a.urlf(format, s...)
	_, _ = fmt.Fprintf(GinkgoWriter, "HTTP PUT: %s\n", url)
	_, _ = fmt.Fprintf(GinkgoWriter, "Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Set("Content-Type", "text/html")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated, http.StatusOK))
}

func (a *App) POSTResponse(data, format string, s ...any) (*http.Response, error) {
	url := a.urlf(format, s...)
	_, _ = fmt.Fprintf(GinkgoWriter, "HTTP POST: %s\n", url)
	_, _ = fmt.Fprintf(GinkgoWriter, "Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(request)
}

func (a *App) POST(data, format string, s ...any) *http.Response {
	response, err := a.POSTResponse(data, format, s...)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated, http.StatusOK, http.StatusAccepted))
	return response
}

func (a *App) DELETEResponse(format string, s ...any) (*http.Response, error) {
	url := a.urlf(format, s...)
	_, _ = fmt.Fprintf(GinkgoWriter, "HTTP DELETE: %s\n", url)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	Expect(err).NotTo(HaveOccurred())
	return http.DefaultClient.Do(request)
}

func (a *App) DELETE(format string, s ...any) {
	response, err := a.DELETEResponse(format, s...)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusGone, http.StatusNoContent, http.StatusOK))
}

func (a *App) GetRawResponse(format string, s ...any) *http.Response {
	url := a.urlf(format, s...)
	_, _ = fmt.Fprintf(GinkgoWriter, "HTTP GET: %s\n", url)
	response, err := http.Get(url)
	Expect(err).NotTo(HaveOccurred())
	return response
}

func (a *App) urlf(format string, s ...any) string {
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
