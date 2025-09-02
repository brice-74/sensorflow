package app

import (
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/internal/kit"
	"github.com/brice-74/sensorflow/pkg/log"
	"github.com/rs/zerolog"
)

func PrepareLogger(conf *Config) (log.FiberLoggerInterface, func() error, error) {
	if conf.Env == config.EnvLocal {
		return kit.Zlog(zerolog.DebugLevel, zerolog.DebugLevel, "./logs/ingestion.log"), func() error { return nil }, nil
	}

	return kit.Sentry(conf.Sentry)
}
