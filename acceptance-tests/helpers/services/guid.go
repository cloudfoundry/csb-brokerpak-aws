package services

import (
	"strings"

	"csbbrokerpakaws/acceptance-tests/helpers/cf"
)

func (s *ServiceInstance) GUID() string {
	if s.guid == "" {
		out, _ := cf.Run("service", s.Name, "--guid")
		s.guid = strings.TrimSpace(out)
	}

	return s.guid
}
