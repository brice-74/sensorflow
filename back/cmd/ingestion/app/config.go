package app

import (
	"github.com/brice-74/sensorflow/internal/config"
	configpkg "github.com/brice-74/sensorflow/pkg/config"
)

type Config struct {
	config.Instance
	config.Env
	config.InfluxDB
	config.Sentry
	config.GRPC
	config.Postgres
	config.Redis
	config.WorkerPool
}

func (c *Config) Define(loader *configpkg.Loader) {
	c.Instance.Define(loader)
	c.Env.Define(loader)
	c.InfluxDB.Define(loader)
	c.Sentry.Define(loader)
	c.GRPC.Define(loader)
	c.Postgres.Define(loader)
	c.Redis.Define(loader)
	c.WorkerPool.Define(loader)
}

func ParseConfig() (*Config, error) {
	cfg := new(Config)
	if err := configpkg.NewLoader(configpkg.Both, cfg).Parse(); err != nil {
		return nil, err
	}
	return cfg, nil
}
