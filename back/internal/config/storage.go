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

type Postgres struct {
	Host, Port, Database, User, Password                 string
	MaxOpenConns, MaxIdleConns                           int
	ConnMaxIdleTime, ConnMaxLifetime, InitialPingTimeout time.Duration
}

func (c *Postgres) Define(loader *config.Loader) {
	loader.String(&c.Host, "postgres_host", "POSTGRES_HOST", "", "Postgres host").
		Required()
	loader.String(&c.Port, "postgres_port", "POSTGRES_PORT", "", "Postgres port").
		Required().
		Validate(validatePort)
	loader.String(&c.Database, "postgres_database", "POSTGRES_DATABASE", "", "Postgres database name").
		Required()
	loader.String(&c.User, "postgres_user", "POSTGRES_USER", "", "Postgres user").
		Required()
	loader.String(&c.Password, "postgres_password", "POSTGRES_PASSWORD", "", "Postgres password").
		Required()
	loader.Int(&c.MaxOpenConns, "postgres_max_open_conns", "POSTGRES_MAX_OPEN_CONNS", 0,
		"Maximum number of open connections to the Postgres database",
	)
	loader.Int(&c.MaxIdleConns, "postgres_max_idle_conns", "POSTGRES_MAX_IDLE_CONNS", 2,
		"Maximum number of idle connections maintained in the pool",
	)
	loader.Duration(&c.ConnMaxIdleTime, "postgres_conn_max_idle_time", "POSTGRES_CONN_MAX_IDLE_TIME", 0,
		"Maximum amount of time a connection can remain idle before being closed",
	)
	loader.Duration(&c.ConnMaxLifetime, "postgres_conn_max_lifetime", "POSTGRES_CONN_MAX_LIFETIME", 0,
		"Maximum total lifetime of a connection before it is closed",
	)
	loader.Duration(&c.InitialPingTimeout, "postgres_initial_ping_timeout", "POSTGRES_INITIAL_PING_TIMEOUT", 0,
		"Maximum duration to wait for the initial ping to the database when opening the connection",
	)
}
