package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/brice-74/sensorflow/pkg/config"
	"github.com/brice-74/sensorflow/pkg/parsestring"
)

type Network struct {
	Port        string
	Keepalive   bool
	IdleTimeout time.Duration
	BodyLimit   int64
	Concurrency int
}

func (c *Network) Define(loader *config.Loader) {
	loader.String(&c.Port, "net_port", "NET_PORT", "", "Exposed network port").
		Required().
		Validate(func(s string) error {
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
		})
	loader.Bool(&c.Keepalive, "net_keep_alive", "NET_KEEPALIVE", true,
		"Check whether the server uses persistent HTTP connections (Keep-Alive)")
	loader.Duration(&c.IdleTimeout, "net_idle_timeout", "NET_IDLE_TIMEOUT", 0,
		"Maximum time a connection can remain inactive before being closed")
	config.AddOption(loader, &c.BodyLimit, "net_body_limit", "NET_BODY_LIMIT", 0,
		"Limits the maximum size of the request body", parsestring.ByteSize)
	config.AddOption(loader, &c.Concurrency, "net_concurrency", "NET_CONCURRENCY", 0,
		"Limits the number of concurrent requests/goroutines", parsestring.EvalCalc)
}

type HTTP struct {
	Network
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

func (c *HTTP) Define(loader *config.Loader) {
	c.Network.Define(loader)
	loader.Duration(&c.WriteTimeout, "http_write_timeout", "HTTP_WRITE_TIMEOUT", 0,
		"Timeout for writing the response")
	loader.Duration(&c.ReadTimeout, "http_read_timeout", "HTTP_READ_TIMEOUT", 0,
		"Timeout for reading the incoming request")
}

type GRPC struct {
	Network
	MaxConnectionAge time.Duration
	MaxRecvMsgSize   int
	MaxSendMsgSize   int
}

func (c *GRPC) Define(loader *config.Loader) {
	c.Network.Define(loader)
	loader.Duration(&c.MaxConnectionAge, "grpc_max_connection_age", "GRPC_MAX_CONNECTION_AGE", 0,
		"Maximum age of a connection before it is closed")
	loader.Int(&c.MaxRecvMsgSize, "grpc_max_recv_msg_size", "GRPC_MAX_RECV_MSG_SIZE", 0,
		"Maximum size in bytes of a received gRPC message")
	loader.Int(&c.MaxSendMsgSize, "grpc_max_send_msg_size", "GRPC_MAX_SEND_MSG_SIZE", 0,
		"Maximum size in bytes of a sent gRPC message")
}
