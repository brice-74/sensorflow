package app

import (
	"github.com/brice-74/sensorflow/internal/config"
	configpkg "github.com/brice-74/sensorflow/pkg/config"
)

type Config struct {
	config.Instance
	config.Env
	config.HTTP
	config.InfluxDB
	config.Sentry
}

func ParseConfig() (*Config, error) {
	cfg := new(Config)

	loader := configpkg.NewLoader(configpkg.Both,
		&cfg.Instance,
		&cfg.Env,
		&cfg.HTTP,
		&cfg.InfluxDB,
		&cfg.Sentry,
	)

	if err := loader.Parse(); err != nil {
		return nil, err
	}
	return cfg, nil
}
