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

	args := []string{"update-service", s.Name}
	if cfg.parameters != "" {
		args = append(args, "-c", cfg.parameters)
	}

	if cfg.plan != "" {
		args = append(args, "-p", cfg.plan)
	}

	switch cf.Version() {
	case cf.VersionV8:
		updateServiceWithWait(s.Name, args)
	default:
		updateServiceWithPoll(s.Name, args)
	}
}

func updateServiceWithWait(name string, args []string) {
	args = append(args, "--wait")
	session := cf.Start(args...)
	Eventually(session, time.Hour).Should(Exit(0), func() string {
		out, _ := cf.Run("service", name)
		return out
	})
}

func updateServiceWithPoll(name string, args []string) {
	session := cf.Start(args...)
	Eventually(session, 5*time.Minute).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", name)
		Expect(out).NotTo(MatchRegexp(`status:\s+update failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+update succeeded`))
}
