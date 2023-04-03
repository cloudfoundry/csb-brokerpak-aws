package apps

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// NewPayload reads an HTTP response body
// It does not close the body
func NewPayload(response *http.Response) Payload {
	body, err := io.ReadAll(response.Body)
	Expect(err).NotTo(HaveOccurred())
	return Payload(body)
}

type Payload string

func (p Payload) ParseInto(receiver any) {
	GinkgoHelper()

	Expect(reflect.ValueOf(receiver).Kind()).To(Equal(reflect.Ptr), "must pass a pointer to the receiver")
	Expect(json.Unmarshal([]byte(p), receiver)).To(Succeed())
}

func (p Payload) String() string {
	return string(p)
}
