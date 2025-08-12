package config

import (
	"time"

	"github.com/brice-74/sensorflow/pkg/config"
)

type InfluxDB struct {
	URL, Token, Database           string
	Timeout, IdleConnectionTimeout time.Duration
	MaxIdleConnections             int
}

func (c *InfluxDB) Define(loader *config.Loader) {
	loader.String(&c.URL, "influxdb_url", "INFLUXDB_URL", "", "influxDB v3 api url").
		Required()
	loader.String(&c.Token, "influxdb_token", "INFLUXDB_TOKEN", "", "influxDB v3 token access").
		Required()
	loader.String(&c.Database, "influxdb_database", "INFLUXDB_DATABSE", "", "influxDB v3 database name").
		Required()
	loader.Duration(&c.Timeout, "influxdb_timeout", "INFLUXDB_TIMEOUT", 10*time.Second, "influxDB v3 request timeout")
	loader.Duration(&c.IdleConnectionTimeout, "influxdb_idle_connection_timeout", "INFLUXDB_IDLE_CONNECTION_TIMEOUT", 90*time.Second, "influxDB v3 idle connection timeout")
	loader.Int(&c.MaxIdleConnections, "influxdb_max_idle_connections", "INFLUXDB_MAX_IDLE_CONNECTIONS", 10, "influxDB v3 max idle connexions")
}
