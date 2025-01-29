package servicekeys

import (
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"encoding/json"
	"reflect"
	"strings"

	. "github.com/onsi/gomega"
)

func (s *ServiceKey) Get(receiver any) {
	Expect(reflect.ValueOf(receiver).Kind()).To(Equal(reflect.Ptr), "receiver must be a pointer")
	out, _ := cf.Run("service-key", s.serviceInstanceName, s.name)

	// The output consists of some text followed by JSON. We are only interested in the JSON
	start := strings.Index(out, "{")
	Expect(start).To(BeNumerically(">", 0), "could not find start of JSON")
	data := []byte(out[start:])

	// Remove the outer '{"credentials":...}' block
	var wrapper struct {
		Credentials any `json:"credentials"`
	}
	err := json.Unmarshal(data, &wrapper)
	Expect(err).NotTo(HaveOccurred())
	data, err = json.Marshal(wrapper.Credentials)
	Expect(err).NotTo(HaveOccurred())

	Expect(json.Unmarshal(data, receiver)).To(Succeed())
}
