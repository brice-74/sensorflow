package config

import (
	"time"

	"github.com/brice-74/sensorflow/pkg/config"
)

type WorkerPool struct {
	Enabled     bool
	MinWorkers  int
	MaxWorkers  int
	QueueSize   int
	IdleTimeout time.Duration
}

func (c *WorkerPool) Define(loader *config.Loader) {
	loader.Bool(&c.Enabled,
		"workerpool_enabled",
		"WORKERPOOL_ENABLED",
		true,
		"Enable or disable the async worker pool",
	)

	loader.Int(&c.MinWorkers,
		"workerpool_min_workers",
		"WORKERPOOL_MIN_WORKERS",
		1,
		"Minimum number of workers",
	)

	loader.Int(&c.MaxWorkers,
		"workerpool_max_workers",
		"WORKERPOOL_MAX_WORKERS",
		10,
		"Maximum number of workers (<=0 means unlimited)",
	)

	loader.Int(&c.QueueSize,
		"workerpool_queue_size",
		"WORKERPOOL_QUEUE_SIZE",
		1000,
		"Maximum number of pending async tasks",
	)

	loader.Duration(&c.IdleTimeout,
		"workerpool_idle_timeout",
		"WORKERPOOL_IDLE_TIMEOUT",
		5*time.Second,
		"Duration after which idle workers are terminated",
	)
}
