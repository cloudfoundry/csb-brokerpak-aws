package bindings

import (
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
)

func (b *Binding) Unbind() {
	cf.Run("unbind-service", b.appName, b.serviceInstanceName)
}
