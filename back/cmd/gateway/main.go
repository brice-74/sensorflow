package main

import (
	"fmt"
	"os"

	"github.com/brice-74/sensorflow/cmd/gateway/app"
)

func main() {
	cfg, err := app.ParseConfig()
	if err != nil {
		fmt.Printf("parse config error: %s", err)
		os.Exit(1)
	}

	logger, flushSentry, err := app.PrepareLogger(cfg)
	if err != nil {
		fmt.Printf("failed to init logger: %s", err)
		os.Exit(1)
	}
	defer flushSentry()

}
