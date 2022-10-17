// Package bindings manages service bindings
package bindings

import (
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
	"encoding/json"

	"github.com/onsi/gomega"
)

type Binding struct {
	name                string
	serviceInstanceName string
	appName             string
}

type config struct {
	bindingName string
	parameters  string
}

type Option func(*config)

func Bind(serviceInstanceName, appName string, opts ...Option) *Binding {
	var c config
	WithOptions(opts...)(&c)

	if c.bindingName == "" {
		c.bindingName = random.Name()
	}

	cmd := []string{
		"bind-service", appName, serviceInstanceName, "--binding-name", c.bindingName,
	}

	if c.parameters != "" {
		cmd = append(cmd, "-c", c.parameters)
	}

	cf.Run(cmd...)
	return &Binding{
		name:                c.bindingName,
		serviceInstanceName: serviceInstanceName,
		appName:             appName,
	}
}

func WithOptions(opts ...Option) Option {
	return func(c *config) {
		for _, o := range opts {
			o(c)
		}
	}
}

func WithName(name string) Option {
	return func(c *config) {
		c.bindingName = name
	}
}

func WithParameters(parameters any) Option {
	return func(c *config) {
		switch p := parameters.(type) {
		case string:
			c.parameters = p
		default:
			params, err := json.Marshal(p)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			c.parameters = string(params)
		}
	}
}
