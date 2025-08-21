package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/brice-74/sensorflow/pkg/config"
	"github.com/brice-74/sensorflow/pkg/parsestring"
)

type HTTP struct {
	Port                                                        string
	GracefulStopTimeout, IdleTimeout, WriteTimeout, ReadTimeout time.Duration
	Keepalive                                                   bool
	BodyLimit                                                   int64
	ConcurrencyLimit                                            int
}

func (c *HTTP) Define(loader *config.Loader) {
	loader.String(&c.Port, "http_port", "HTTP_PORT", "", "Exposed HTTP port").
		Required().
		Validate(validatePort)
	loader.Duration(&c.GracefulStopTimeout, "http_graceful_stop_timeout", "HTTP_GRACEFUL_STOP_TIMEOUT", 10*time.Second, "Maximum time for the server to shut down gracefully otherwise it will shut down abruptly")
	loader.Bool(&c.Keepalive, "http_keep_alive", "HTTP_KEEPALIVE", true, "Enable HTTP persistent connections (Keep-Alive)")
	loader.Duration(&c.IdleTimeout, "http_idle_timeout", "HTTP_IDLE_TIMEOUT", 0, "Maximum time a connection can remain inactive before being closed")
	loader.Duration(&c.WriteTimeout, "http_write_timeout", "HTTP_WRITE_TIMEOUT", 0, "Timeout for writing the response")
	loader.Duration(&c.ReadTimeout, "http_read_timeout", "HTTP_READ_TIMEOUT", 0, "Timeout for reading the incoming request")
	config.AddOption(loader, &c.BodyLimit, "http_body_limit", "HTTP_BODY_LIMIT", 0, "Maximum size of the request body", parsestring.ByteSize)
	config.AddOption(loader, &c.ConcurrencyLimit, "http_concurrency", "HTTP_CONCURRENCY", 0, "Maximum number of concurrent requests/goroutines", parsestring.EvalCalc)
}

type GRPC struct {
	Port string
	GracefulStopTimeout, IdleTimeout, MaxConnectionAge, MaxConnectionAgeGrace,
	KeepaliveTime, KeepaliveTimeout, MinTimeBetweenPings time.Duration
	MaxRecvMsgSize             int
	MaxSendMsgSize             int
	AllowPingWithoutActiveRPCs bool
	MaxConcurrentStreams       uint32
}

func (c *GRPC) Define(loader *config.Loader) {
	loader.String(&c.Port, "grpc_port", "GRPC_PORT", "", "Exposed gRPC port").
		Required().
		Validate(validatePort)
	loader.Duration(&c.GracefulStopTimeout, "grpc_graceful_stop_timeout", "GRPC_GRACEFUL_STOP_TIMEOUT", 10*time.Second, "Maximum time for the server to shut down gracefully otherwise it will shut down abruptly")
	loader.Duration(&c.IdleTimeout, "grpc_idle_timeout", "GRPC_IDLE_TIMEOUT", 0, "Maximum time a connection can remain idle before being closed")
	loader.Duration(&c.MaxConnectionAge, "grpc_max_connection_age", "GRPC_MAX_CONNECTION_AGE", 0, "Maximum age of a connection before it is closed")
	loader.Duration(&c.MaxConnectionAgeGrace, "grpc_max_connection_age_grace", "GRPC_MAX_CONNECTION_AGE_GRACE", 0, "Additive grace period after MaxConnectionAge for finishing in-flight RPCs")
	loader.Int(&c.MaxRecvMsgSize, "grpc_max_recv_msg_size", "GRPC_MAX_RECV_MSG_SIZE", 0, "Maximum size in bytes of a received gRPC message")
	loader.Int(&c.MaxSendMsgSize, "grpc_max_send_msg_size", "GRPC_MAX_SEND_MSG_SIZE", 0, "Maximum size in bytes of a sent gRPC message")
	loader.Duration(&c.MinTimeBetweenPings, "grpc_min_time_between_pings", "GRPC_MIN_TIME_BETWEEN_PINGS", 0, "Minimum duration a client must wait before sending a keepalive ping")
	loader.Bool(&c.AllowPingWithoutActiveRPCs, "grpc_allow_ping_without_active_rpcs", "GRPC_ALLOW_PING_WITHOUT_ACTIVE_RPCS", false, "Allow keepalive pings even when there are no active streams")
}

func validatePort(s string) error {
	if s == "" {
		return fmt.Errorf("port cannot be empty")
	}
	portNum, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}
