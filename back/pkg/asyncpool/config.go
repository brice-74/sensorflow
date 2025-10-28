package asyncpool

import "time"

type Option func(*WorkerPool)

func WithMinWorkers(min int) Option {
	return func(p *WorkerPool) { p.minWorkers = min }
}

func WithMaxWorkers(max int) Option {
	return func(p *WorkerPool) { p.maxWorkers = max }
}

func WithIdleTimeout(d time.Duration) Option {
	return func(p *WorkerPool) { p.idleTimeout = d }
}

func WithQueueSize(size int) Option {
	return func(p *WorkerPool) { p.queueSize = size }
}

func WithPanicHandler(handler PanicHandler) Option {
	return func(p *WorkerPool) { p.panicHandler = handler }
}
