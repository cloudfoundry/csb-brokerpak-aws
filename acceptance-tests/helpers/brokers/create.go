package brokers

import (
	"csbbrokerpakaws/acceptance-tests/helpers/testpath"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/onsi/gomega"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/brokerpaks"
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"csbbrokerpakaws/acceptance-tests/helpers/random"
)

type Option func(broker *Broker)

func Create(opts ...Option) *Broker {
	broker := defaultConfig(opts...)

	brokerApp := apps.Push(
		apps.WithName(broker.Name),
		apps.WithDir(broker.dir),
		apps.WithManifest(newManifest(
			withName(broker.Name),
			withEnv(broker.env()...),
		)),
	)

	schemaName := strings.ReplaceAll(broker.Name, "-", "_")
	cf.Run("bind-service", broker.Name, "csb-sql", "-c", fmt.Sprintf(`{"schema":"%s"}`, schemaName))

	brokerApp.Start()

	cf.Run("create-service-broker", broker.Name, broker.username, broker.password, brokerApp.URL, "--space-scoped")

	broker.app = brokerApp
	return &broker
}

func WithOptions(opts ...Option) Option {
	return func(b *Broker) {
		for _, o := range opts {
			o(b)
		}
	}
}

func WithName(name string) Option {
	return func(b *Broker) {
		b.Name = name
	}
}

func WithPrefix(prefix string) Option {
	return func(b *Broker) {
		b.Name = random.Name(random.WithPrefix(prefix))
	}
}

func WithSourceDir(dir string) Option {
	gomega.Expect(filepath.Join(dir, "cloud-service-broker")).To(gomega.BeAnExistingFile())
	return func(b *Broker) {
		b.dir = dir
	}
}

func WithEnv(env ...apps.EnvVar) Option {
	return func(b *Broker) {
		b.envExtras = append(b.envExtras, env...)
	}
}

func WithReleaseEnv() Option {
	return func(b *Broker) {
		b.envExtras = append(b.envExtras, b.releaseEnv()...)
	}
}

func WithLegacyMySQLEnvFor140() Option {
	return func(b *Broker) {
		if brokerpaks.DetectBrokerpakV140(b.dir) {
			b.envExtras = append(b.envExtras, b.legacyMysqlEnv()...)
		}
	}
}

func WithLatestEnv() Option {
	return func(b *Broker) {
		b.envExtras = append(b.envExtras, b.latestEnv()...)
	}
}

func WithUsername(username string) Option {
	return func(b *Broker) {
		b.username = username
	}
}

func WithPassword(password string) Option {
	return func(b *Broker) {
		b.password = password
	}
}

func defaultConfig(opts ...Option) (broker Broker) {
	defaults := []Option{
		WithName(random.Name(random.WithPrefix("broker"))),
		WithSourceDir(testpath.BrokerpakRoot()),
		WithUsername(random.Name()),
		WithPassword(random.Password()),
		WithEncryptionSecret(random.Password()),
	}
	WithOptions(append(defaults, opts...)...)(&broker)
	return broker
}
