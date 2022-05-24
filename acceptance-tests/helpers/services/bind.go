package services

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/bindings"
)

func (s *ServiceInstance) Bind(app *apps.App) *bindings.Binding {
	return bindings.Bind(s.Name, app.Name)
}
