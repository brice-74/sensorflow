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
	Host, Port, Database, User, Password string
	MaxOpenConns, MaxIdleConns           int
	ConnMaxIdleTime, ConnMaxLifetime     time.Duration
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
}

type Redis struct {
	// Base redis.Options fields
	Network, Addr, ClientName                  string
	Username, Password                         string
	DB, MaxRetries                             int
	MinRetryBackoff, MaxRetryBackoff           time.Duration
	DialTimeout, ReadTimeout, WriteTimeout     time.Duration
	ContextTimeoutEnabled                      bool
	ReadBufferSize, WriteBufferSize            int
	PoolFIFO                                   bool
	PoolSize                                   int
	PoolTimeout                                time.Duration
	MinIdleConns, MaxIdleConns, MaxActiveConns int
	ConnMaxIdleTime, ConnMaxLifetime           time.Duration

	// HealthyClient wrapper options
	HealthPingTimeout                      time.Duration
	HealthInitialBackoff, HealthMaxBackoff time.Duration
	HealthMultiplier                       float64
	HealthJitterPct                        float64
}

func (c *Redis) Define(loader *config.Loader) {
	loader.String(&c.Network, "redis_network", "REDIS_NETWORK", "tcp", "Redis network mode (tcp or unix)")
	loader.String(&c.Addr, "redis_addr", "REDIS_ADDR", "", "Redis address host:port").Required()

	loader.String(&c.ClientName, "redis_client_name", "REDIS_CLIENT_NAME", "", "Redis client name")

	loader.String(&c.Username, "redis_username", "REDIS_USERNAME", "", "Redis ACL username")
	loader.String(&c.Password, "redis_password", "REDIS_PASSWORD", "", "Redis ACL password")

	loader.Int(&c.DB, "redis_db", "REDIS_DB", 0, "Redis database number")

	// Retry logic
	loader.Int(&c.MaxRetries, "redis_max_retries", "REDIS_MAX_RETRIES", 3, "Redis max retries before giving up")
	loader.Duration(&c.MinRetryBackoff, "redis_min_retry_backoff", "REDIS_MIN_RETRY_BACKOFF",
		8*time.Millisecond, "Redis minimum retry backoff")
	loader.Duration(&c.MaxRetryBackoff, "redis_max_retry_backoff", "REDIS_MAX_RETRY_BACKOFF",
		512*time.Millisecond, "Redis maximum retry backoff")

	// Timeouts
	loader.Duration(&c.DialTimeout, "redis_dial_timeout", "REDIS_DIAL_TIMEOUT",
		5*time.Second, "Redis dial timeout")
	loader.Duration(&c.ReadTimeout, "redis_read_timeout", "REDIS_READ_TIMEOUT",
		3*time.Second, "Redis read timeout")
	loader.Duration(&c.WriteTimeout, "redis_write_timeout", "REDIS_WRITE_TIMEOUT",
		3*time.Second, "Redis write timeout")

	loader.Bool(&c.ContextTimeoutEnabled, "redis_context_timeout_enabled",
		"REDIS_CONTEXT_TIMEOUT_ENABLED", true, "Enable context timeout handling")

	// Buffers
	loader.Int(&c.ReadBufferSize, "redis_read_buffer_size", "REDIS_READ_BUFFER_SIZE",
		32768, "Redis read buffer size")
	loader.Int(&c.WriteBufferSize, "redis_write_buffer_size", "REDIS_WRITE_BUFFER_SIZE",
		32768, "Redis write buffer size")

	// Pooling
	loader.Bool(&c.PoolFIFO, "redis_pool_fifo", "REDIS_POOL_FIFO", false, "Use FIFO instead of LIFO pool")
	loader.Int(&c.PoolSize, "redis_pool_size", "REDIS_POOL_SIZE", 0, "Redis pool size")
	loader.Duration(&c.PoolTimeout, "redis_pool_timeout", "REDIS_POOL_TIMEOUT",
		0, "Redis pool timeout if all connections are busy")

	loader.Int(&c.MinIdleConns, "redis_min_idle_conns", "REDIS_MIN_IDLE_CONNS", 0, "Redis minimum idle conns")
	loader.Int(&c.MaxIdleConns, "redis_max_idle_conns", "REDIS_MAX_IDLE_CONNS", 0, "Redis maximum idle conns")
	loader.Int(&c.MaxActiveConns, "redis_max_active_conns", "REDIS_MAX_ACTIVE_CONNS", 0, "Redis maximum active conns")

	loader.Duration(&c.ConnMaxIdleTime, "redis_conn_max_idle_time",
		"REDIS_CONN_MAX_IDLE_TIME", 30*time.Minute, "Max idle time for redis connections")
	loader.Duration(&c.ConnMaxLifetime, "redis_conn_max_lifetime",
		"REDIS_CONN_MAX_LIFETIME", 0, "Max lifetime for redis connections")

	// HealthyClient (wrapper)
	loader.Duration(&c.HealthPingTimeout, "redis_health_ping_timeout", "REDIS_HEALTH_PING_TIMEOUT",
		500*time.Millisecond, "Ping timeout for redis health check")
	loader.Duration(&c.HealthInitialBackoff, "redis_health_initial_backoff",
		"REDIS_HEALTH_INITIAL_BACKOFF", 100*time.Millisecond, "Initial backoff for redis recovery")
	loader.Duration(&c.HealthMaxBackoff, "redis_health_max_backoff",
		"REDIS_HEALTH_MAX_BACKOFF", 30*time.Second, "Max backoff for redis recovery")

	loader.Float64(&c.HealthMultiplier, "redis_health_multiplier",
		"REDIS_HEALTH_MULTIPLIER", 2, "Backoff multiplier for redis recovery")
	loader.Float64(&c.HealthJitterPct, "redis_health_jitter_pct",
		"REDIS_HEALTH_JITTER_PCT", 0.0, "Jitter percentage for redis recovery")
}
