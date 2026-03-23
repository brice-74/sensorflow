package ingestor

import (
	"time"

	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/pkg/errors"
)

type Config[T any] struct {
	Name   string
	Logger log.Logger
	Levels []IngestionLevel[T]

	FlushBatchSize int

	Tick               time.Duration
	FlushDelay         time.Duration
	RetryDelay         time.Duration
	LevelProbeInterval time.Duration

	MaxBufferBeforeFallback int
}

func defaultConfig[T any]() Config[T] {
	return Config[T]{
		Tick:                    time.Second,
		FlushBatchSize:          100,
		FlushDelay:              2 * time.Second,
		RetryDelay:              5 * time.Second,
		MaxBufferBeforeFallback: 1000,
		LevelProbeInterval:      10 * time.Second,
	}
}

func (cfg *Config[T]) withDefaults() {
	def := defaultConfig[T]()

	if cfg.Tick == 0 {
		cfg.Tick = def.Tick
	}
	if cfg.FlushBatchSize == 0 {
		cfg.FlushBatchSize = def.FlushBatchSize
	}
	if cfg.FlushDelay == 0 {
		cfg.FlushDelay = def.FlushDelay
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = def.RetryDelay
	}
	if cfg.MaxBufferBeforeFallback == 0 {
		cfg.MaxBufferBeforeFallback = def.MaxBufferBeforeFallback
	}
	if cfg.LevelProbeInterval == 0 {
		cfg.LevelProbeInterval = def.LevelProbeInterval
	}
}

func (cfg *Config[T]) validate() error {
	if cfg.Name == "" {
		return errors.WrapMsg("name is required")
	}
	if cfg.Logger == nil {
		return errors.WrapMsg("logger is required")
	}
	if len(cfg.Levels) == 0 {
		return errors.WrapMsg("at least one level is required")
	}
	return nil
}
