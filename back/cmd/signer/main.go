package main

import (
	"fmt"
	"os"

	"github.com/brice-74/sensorflow/cmd/signer/app"
	"github.com/brice-74/sensorflow/internal/log"
)

func main() {
	cfg, err := app.ParseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse config error:", err)
		os.Exit(1)
	}

	logger, err := app.PrepareLogger(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		os.Exit(1)
	}
	logger = logger.With(
		log.Tags{"api_identifier": cfg.Instance.Identifier},
	).(log.Fiber)

	signerService, err := app.NewService(cfg.Signer, logger)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	if err := app.ServeHTTP(cfg.Signer, app.HTTPDeps{
		Logger:  logger,
		Service: signerService,
	}); err != nil {
		logger.Error(err, log.Tags{"server": "closed"})
	}
}
