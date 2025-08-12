package influx

import (
	"fmt"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/brice-74/sensorflow/internal/config"
)

func NewClient(cfg config.InfluxDB) (*influxdb3.Client, error) {
	config := influxdb3.ClientConfig{
		Host:                  cfg.URL,
		Token:                 cfg.Token,
		Database:              cfg.Database,
		Timeout:               cfg.Timeout,
		IdleConnectionTimeout: cfg.IdleConnectionTimeout,
		MaxIdleConnections:    cfg.MaxIdleConnections,
	}
	client, err := influxdb3.New(config)
	if err != nil {
		return nil, fmt.Errorf("error on new influx client, error: %s", err)
	}

	return client, nil
}
