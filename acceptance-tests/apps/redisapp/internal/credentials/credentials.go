package credentials

import (
	"crypto/tls"
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

type Credentials struct {
	Host           string `mapstructure:"host"`
	ReaderEndpoint string `mapstructure:"reader_endpoint"`
	Password       string `mapstructure:"password"`
	TLSPort        int    `mapstructure:"tls_port"`
}

func Read() (Credentials, error) {
	app, err := cfenv.Current()
	if err != nil {
		return Credentials{}, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("redis")
	if err != nil {
		return Credentials{}, fmt.Errorf("error reading Redis service details")
	}

	var r Credentials
	if err := mapstructure.Decode(svs[0].Credentials, &r); err != nil {
		return Credentials{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.Host == "" || r.Password == "" || r.TLSPort == 0 {
		return Credentials{}, fmt.Errorf("parsed credentials are not valid")
	}

	return r, nil
}

func (c Credentials) Client() *redis.Client {
	return c.client(c.Host)
}

func (c Credentials) ReaderClient() (*redis.Client, error) {
	if c.ReaderEndpoint == "" {
		return nil, fmt.Errorf("no reader endpoint in the credentials")
	}
	return c.client(c.ReaderEndpoint), nil
}

func (c Credentials) client(endpoint string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%d", endpoint, c.TLSPort),
		Password:  c.Password,
		DB:        0,
		TLSConfig: &tls.Config{},
	})
}
