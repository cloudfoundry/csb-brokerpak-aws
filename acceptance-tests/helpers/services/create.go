// Package services manages service instances
package services

import (
	"encoding/json"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"csbbrokerpakaws/acceptance-tests/helpers/brokers"
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
)

type ServiceInstance struct {
	Name string
	guid string
}

type config struct {
	name              string
	serviceBrokerName func() string
	parameters        string
	plan              string
}

type Option func(*config)

func CreateInstance(offering string, opts ...Option) *ServiceInstance {
	cfg := defaultConfig(offering, opts...)
	Expect(cfg.plan).ToNot(BeEmpty())
	args := []string{
		"create-service",
		offering,
		cfg.plan,
		cfg.name,
		"-b",
		cfg.serviceBrokerName(),
	}

	if cfg.parameters != "" {
		args = append(args, "-c", cfg.parameters)
	}

	switch cf.Version() {
	case cf.VersionV8:
		createInstanceWithWait(cfg.name, args)
	default:
		createInstanceWithPoll(cfg.name, args)
	}

	return &ServiceInstance{Name: cfg.name}
}

func createInstanceWithWait(name string, args []string) {
	args = append(args, "--wait")
	session := cf.Start(args...)
	Eventually(session, time.Hour).Should(Exit(0), func() string {
		out, _ := cf.Run("service", name)
		return out
	})
}

func createInstanceWithPoll(name string, args []string) {
	session := cf.Start(args...)
	Eventually(session, 5*time.Minute).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", name)
		Expect(out).NotTo(MatchRegexp(`status:\s+create failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+create succeeded`))
}

func WithDefaultBroker() Option {
	return func(c *config) {
		c.serviceBrokerName = brokers.DefaultBrokerName
	}
}

func WithBroker(broker *brokers.Broker) Option {
	return func(c *config) {
		c.serviceBrokerName = func() string { return broker.Name }
	}
}

func WithBrokerName(name string) Option {
	return func(c *config) {
		c.serviceBrokerName = func() string { return name }
	}
}

func WithParameters(parameters any) Option {
	return func(c *config) {
		switch p := parameters.(type) {
		case string:
			c.parameters = p
		default:
			params, err := json.Marshal(p)
			Expect(err).NotTo(HaveOccurred())
			c.parameters = string(params)
		}
	}
}

func WithPlan(plan string) Option {
	return func(c *config) {
		c.plan = plan
	}
}

func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

func WithOptions(opts ...Option) Option {
	return func(c *config) {
		for _, o := range opts {
			o(c)
		}
	}
}

func defaultConfig(offering string, opts ...Option) config {
	var cfg config
	WithOptions(append([]Option{
		WithDefaultBroker(),
		WithName(random.Name(random.WithPrefix(offering))),
	}, opts...)...)(&cfg)
	return cfg
}
