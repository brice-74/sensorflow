package app

import (
	"fmt"
	"net"
	"time"

	"github.com/brice-74/sensorflow/internal/config"
	configpkg "github.com/brice-74/sensorflow/pkg/config"
)

type Signer struct {
	Addr                string
	CertDir             string
	CACommonName        string
	ServerCommonName    string
	ServerDNSNames      []string
	CADays              int
	ServerDays          int
	ClientDays          int
	ReadHeaderTimeout   time.Duration
	WriteTimeout        time.Duration
	GracefulStopTimeout time.Duration
}

func (c *Signer) Define(loader *configpkg.Loader) {
	l := loader.AddPrefix("signer", "SIGNER")

	isGreaterThan0 := config.IsGreaterThan(0)

	l.String(&c.Addr, "addr", "ADDR", ":8080", "signer listen address").
		Validate(func(v string) error {
			if _, err := net.ResolveTCPAddr("tcp", v); err != nil {
				return fmt.Errorf("invalid listen addr: %w", err)
			}
			return nil
		})
	l.String(&c.CertDir, "cert_dir", "CERT_DIR", "/certs", "certificate directory")
	l.String(&c.CACommonName, "ca_common_name", "CA_COMMON_NAME", "SensorflowCA", "CA certificate common name")
	l.String(&c.ServerCommonName, "server_common_name", "SERVER_COMMON_NAME", "nginx.local", "server certificate common name")
	l.StringSlice(&c.ServerDNSNames, "server_dns_names", "SERVER_DNS_NAMES", []string{"nginx", "nginx.local", "localhost"}, "server certificate DNS SANs")
	l.Int(&c.CADays, "ca_days", "CA_DAYS", 3650, "CA validity in days").
		Validate(isGreaterThan0)
	l.Int(&c.ServerDays, "server_days", "SERVER_DAYS", 365, "server cert validity in days").
		Validate(isGreaterThan0)
	l.Int(&c.ClientDays, "client_days", "CLIENT_DAYS", 90, "client cert max validity in days").
		Validate(isGreaterThan0)
	l.Duration(&c.ReadHeaderTimeout, "read_header_timeout", "READ_HEADER_TIMEOUT", 5*time.Second, "HTTP read header timeout")
	l.Duration(&c.WriteTimeout, "write_timeout", "WRITE_TIMEOUT", 30*time.Second, "HTTP write timeout")
	l.Duration(&c.GracefulStopTimeout, "graceful_stop_timeout", "GRACEFUL_STOP_TIMEOUT", 15*time.Second, "graceful server shutdown timeout")
}

type Config struct {
	config.Instance
	config.Env
	Signer
}

func (c *Config) Define(loader *configpkg.Loader) {
	c.Instance.Define(loader)
	c.Env.Define(loader)
	c.Signer.Define(loader)
}

func ParseConfig() (*Config, error) {
	cfg := new(Config)
	if err := configpkg.NewLoader(configpkg.Both, cfg).Parse(); err != nil {
		return nil, err
	}
	return cfg, nil
}
