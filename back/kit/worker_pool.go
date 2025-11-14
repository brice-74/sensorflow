package kit

import (
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/pkg/asyncpool"
)

func NewWorkerPoolFromConfig(cfg *config.WorkerPool, panicHandler asyncpool.PanicHandler) *asyncpool.WorkerPool {
	if !cfg.Enabled {
		return nil
	}

	return asyncpool.NewWorkerPool(
		asyncpool.WithMinWorkers(cfg.MinWorkers),
		asyncpool.WithMaxWorkers(cfg.MaxWorkers),
		asyncpool.WithQueueSize(cfg.QueueSize),
		asyncpool.WithIdleTimeout(cfg.IdleTimeout),
		asyncpool.WithPanicHandler(panicHandler),
	)
}
