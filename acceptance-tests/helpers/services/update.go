package services

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"csbbrokerpakaws/acceptance-tests/helpers/cf"
)

func (s *ServiceInstance) Update(opts ...Option) {
	var cfg config
	WithOptions(opts...)(&cfg)

	args := []string{"update-service", s.Name, "--wait"}
	if cfg.parameters != "" {
		args = append(args, "-c", cfg.parameters)
	}

	if cfg.plan != "" {
		args = append(args, "-p", cfg.plan)
	}

	session := cf.Start(args...)
	Eventually(session, time.Hour).Should(Exit(0), func() string {
		out, _ := cf.Run("service", cfg.name)
		return out
	})
}
