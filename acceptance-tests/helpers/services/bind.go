package services

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/bindings"
)

type bindConfig struct {
	parameters any
	name       string
}

type BindOption func(*bindConfig)

func (s *ServiceInstance) Bind(app *apps.App, opts ...BindOption) *bindings.Binding {
	var c bindConfig
	WithBindOptions(opts...)(&c)

	var bo []bindings.Option
	if c.parameters != nil {
		bo = append(bo, bindings.WithParameters(c.parameters))
	}
	if c.name != "" {
		bo = append(bo, bindings.WithName(c.name))
	}

	return bindings.Bind(s.Name, app.Name, bo...)
}

func WithBindOptions(opts ...BindOption) BindOption {
	return func(c *bindConfig) {
		for _, o := range opts {
			o(c)
		}
	}
}

func WithBindParameters(params any) BindOption {
	return func(c *bindConfig) {
		c.parameters = params
	}
}

func WithBindingName(name string) BindOption {
	return func(c *bindConfig) {
		c.name = name
	}
}
