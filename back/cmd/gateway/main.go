package main

import (
	"fmt"
	"os"

	"github.com/brice-74/sensorflow/cmd/gateway/app"
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

	if err := app.RunServer(logger, cfg.HTTP.Port); err != nil {
		logger.Error(err, log.Tags{"server": "closed"})
	}
}
