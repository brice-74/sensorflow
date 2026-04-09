package app

import (
	"github.com/brice-74/sensorflow/internal/config"
	configpkg "github.com/brice-74/sensorflow/pkg/config"
)

type Config struct {
	config.Instance
	config.Env
	config.HTTP
	config.Sentry
}

func (c *Config) Define(loader *configpkg.Loader) {
	c.Instance.Define(loader)
	c.Env.Define(loader)
	c.HTTP.Define(loader)
	c.Sentry.Define(loader)
}

func ParseConfig() (*Config, error) {
	cfg := new(Config)
	if err := configpkg.NewLoader(configpkg.Both, cfg).Parse(); err != nil {
		return nil, err
	}
	return cfg, nil
}
