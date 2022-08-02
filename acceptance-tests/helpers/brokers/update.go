package brokers

import (
	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
)

func (b *Broker) UpdateBroker(dir string) {
	WithEnv(b.latestEnv()...)(b)

	b.app.Push(
		apps.WithName(b.Name),
		apps.WithDir(dir),
		apps.WithStartedState(),
		apps.WithManifest(newManifest(
			withName(b.Name),
			withEnv(b.env()...),
		)),
	)

	cf.Run("update-service-broker", b.Name, b.username, b.password, b.app.URL)
}

func (b *Broker) UpdateEnv(env ...apps.EnvVar) {
	WithEnv(env...)(b)
	b.app.SetEnv(b.env()...)
	b.app.Restart()

	cf.Run("update-service-broker", b.Name, b.username, b.password, b.app.URL)
}

func (b *Broker) UpdateEncryptionSecrets(secrets ...EncryptionSecret) {
	WithEncryptionSecrets(secrets...)
	b.app.SetEnv(b.env()...)

	cf.Run("update-service-broker", b.Name, b.username, b.password, b.app.URL)
}
