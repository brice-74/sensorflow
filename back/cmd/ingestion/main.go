package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/brice-74/sensorflow/cmd/ingestion/app"
	"github.com/brice-74/sensorflow/internal/adapters/influx"
	"github.com/brice-74/sensorflow/pkg/log"
)

func main() {
	cfg, err := app.ParseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse config error:", err)
		os.Exit(1)
	}

	logger, flushSentry, err := app.PrepareLogger(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		os.Exit(1)
	}
	defer flushSentry()

	logger = logger.With(
		log.Tags{"api_identifier": cfg.Instance.Identifier},
	).(log.FiberLoggerInterface)

	client, err := influx.NewClient(cfg.InfluxDB)
	if err != nil {
		fmt.Println("failed open cleint", err)
	}

	p1 := influxdb3.NewPoint("stat",
		map[string]string{
			"location": "Paris",
		},
		map[string]any{
			"temperature": 24.5,
			"humidity":    40,
		},
		time.Now(),
	)

	points := []*influxdb3.Point{p1}

	if err := client.WritePoints(context.Background(), points); err != nil {
		fmt.Println("failed send", err)
	}

	time.Sleep(1 * time.Second)
}
