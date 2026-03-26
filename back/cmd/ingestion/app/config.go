package app

import (
	"github.com/brice-74/sensorflow/internal/config"
	configpkg "github.com/brice-74/sensorflow/pkg/config"
)

type OrchestrtorsAsyncTasksConfig struct{ config.WorkerPool }

func (c *OrchestrtorsAsyncTasksConfig) Define(loader *configpkg.Loader) {
	l := loader.AddPrefix("orchestrators_asynctasks", "ORCHESTRATORS_ASYNCTASKS")
	c.WorkerPool.Define(l)
}

type Config struct {
	config.Instance
	config.Env
	config.Sentry
	config.GRPC
	config.Postgres
	config.Redis
	config.Clickhouse
	OrchestrtorsAsyncTasksConfig
}

func (c *Config) Define(loader *configpkg.Loader) {
	c.Instance.Define(loader)
	c.Env.Define(loader)
	c.Sentry.Define(loader)
	c.GRPC.Define(loader)
	c.Postgres.Define(loader)
	c.Redis.Define(loader)
	c.Clickhouse.Define(loader)
	c.OrchestrtorsAsyncTasksConfig.Define(loader)
}

func ParseConfig() (*Config, error) {
	cfg := new(Config)
	if err := configpkg.NewLoader(configpkg.Both, cfg).Parse(); err != nil {
		return nil, err
	}
	return cfg, nil
}
