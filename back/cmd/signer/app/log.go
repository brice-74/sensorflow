package app

import (
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/kit"
	"github.com/rs/zerolog"
)

func PrepareLogger(*Config) (log.Fiber, error) {
	return kit.Zlog(zerolog.DebugLevel, zerolog.DebugLevel, "./logs/signer.log"), nil
}
