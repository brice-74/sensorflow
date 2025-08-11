package app

import (
	"github.com/brice-74/sensorflow/internal/config"
	configpkg "github.com/brice-74/sensorflow/pkg/config"
)

type Config struct {
	config.Env
	config.InfluxDB
	config.Sentry
}

func (c *Config) Parse(loader *configpkg.Loader) error {
	c.Env.Define(loader)
	c.InfluxDB.Define(loader)
	c.Sentry.Define(loader)
	return loader.Parse()
}

func ParseConfig() (*Config, error) {
	cfg := new(Config)
	loader := configpkg.NewLoader(configpkg.Both)
	if err := cfg.Parse(loader); err != nil {
		return nil, err
	}
	return cfg, nil
}
