package brokers

import "csbbrokerpakaws/acceptance-tests/helpers/apps"

type Broker struct {
	Name      string
	username  string
	password  string
	secrets   []EncryptionSecret
	dir       string
	envExtras []apps.EnvVar
	app       *apps.App
}
