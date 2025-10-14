package app

import (
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/kit"

	"github.com/rs/zerolog"
)

func PrepareLogger(conf *Config) (log.Fiber, func() error, error) {
	if conf.Env == config.EnvLocal {
		return kit.Zlog(zerolog.DebugLevel, zerolog.DebugLevel, "./logs/ingestion.log"), func() error { return nil }, nil
	}

	return kit.Sentry(conf.Sentry)
}
