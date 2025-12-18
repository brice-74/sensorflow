package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool interface {
	Submit(task Task) error
}

// PanicHandler is an optional callback for recovered panics in tasks.
type PanicHandler = func(any)

var (
	ErrQueueClosed     = errors.New("asyncpool queue closed")
	ErrQueueFull       = errors.New("asyncpool queue full")
	ErrShutdownTimeout = errors.New("asyncpool shutdown timeout")
)

// BoundedWorkerPool is a lightweight, adaptive worker pool for executing tasks asynchronously
// in order of submission (FIFO) and dynamically adjusts the number of workers.
type BoundedWorkerPool struct {
	queue         chan Task     // Channel holding pending tasks
	minWorkers    int           // Minimum number of workers (can be 0)
	maxWorkers    int           // Maximum workers (<=0 means unlimited)
	idleTimeout   time.Duration // Time after which idle workers exit
	activeWorkers atomic.Int64  // Current active workers
	queueSize     int           // Size of the task queue
	panicHandler  PanicHandler  // Optional panic handler

	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	isClosed atomic.Bool
}

func NewBoundedWorkerPool(opts ...BoundedWorkerPoolOption) *BoundedWorkerPool {
	p := &BoundedWorkerPool{
		minWorkers:  1,
		maxWorkers:  10,
		idleTimeout: 5 * time.Second,
		queueSize:   1000,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.queue = make(chan Task, p.queueSize)
	p.ctx, p.cancel = context.WithCancel(context.Background())

	for i := 0; i < p.minWorkers; i++ {
		p.spawnWorker()
	}

	return p
}

// Submit enqueues a task to the pool. Returns ErrQueueClosed or ErrQueueFull.
func (p *BoundedWorkerPool) Submit(task Task) error {
	if p.isClosed.Load() {
		return ErrQueueClosed
	}

	select {
	case p.queue <- task:
		p.tryScale()
	default:
		return ErrQueueFull
	}
	return nil
}

// spawnWorker starts a new worker goroutine consuming tasks from the queue.
func (p *BoundedWorkerPool) spawnWorker() {
	p.activeWorkers.Add(1)
	p.wg.Add(1)

	go func() {
		defer func() {
			p.activeWorkers.Add(-1)
			p.wg.Done()
		}()

		idleTimer := time.NewTimer(p.idleTimeout)
		defer idleTimer.Stop()

		for {
			select {
			case <-p.ctx.Done():
				return
			case task, ok := <-p.queue:
				if !ok {
					return
				}
				func() {
					defer func() {
						if r := recover(); r != nil && p.panicHandler != nil {
							p.panicHandler(r)
						}
					}()
					task.Do()
				}()
				if !idleTimer.Stop() {
					<-idleTimer.C
				}
				idleTimer.Reset(p.idleTimeout)
			case <-idleTimer.C:
				if p.activeWorkers.Load() > int64(p.minWorkers) {
					return
				}
				idleTimer.Reset(p.idleTimeout)
			}
		}
	}()
}

// tryScale spawns a new worker if there are pending tasks and we haven't reached maxWorkers.
func (p *BoundedWorkerPool) tryScale() {
	qLen := len(p.queue)
	active := int(p.activeWorkers.Load())

	if qLen > 0 && (p.maxWorkers <= 0 || active < p.maxWorkers) {
		p.spawnWorker()
	}
}

// Shutdown gracefully stops the pool, waiting for all current and queued tasks to complete.
// Timeout can be provided via context.
func (p *BoundedWorkerPool) Shutdown(ctx context.Context) error {
	if !p.isClosed.Swap(true) {
		close(p.queue)
	}

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ErrShutdownTimeout
	case <-done:
		return nil
	}
}

// ForceShutdown stops the pool immediately and drains pending tasks still queued.
// Tasks that are currently executing are not forcibly .
func (p *BoundedWorkerPool) ForceShutdown() {
	if !p.isClosed.Swap(true) {
		close(p.queue)
		p.cancel()
	}

	for range p.queue {
	}
	p.wg.Wait()
}

// ActiveWorkers returns the current number of active workers.
func (p *BoundedWorkerPool) ActiveWorkers() int64 {
	return p.activeWorkers.Load()
}

// QueueLength returns the current number of tasks waiting in the queue.
func (p *BoundedWorkerPool) QueueLength() int {
	return len(p.queue)
}

type BoundedWorkerPoolOption func(*BoundedWorkerPool)

func WithMinWorkers(min int) BoundedWorkerPoolOption {
	return func(p *BoundedWorkerPool) { p.minWorkers = min }
}

func WithMaxWorkers(max int) BoundedWorkerPoolOption {
	return func(p *BoundedWorkerPool) { p.maxWorkers = max }
}

func WithIdleTimeout(d time.Duration) BoundedWorkerPoolOption {
	return func(p *BoundedWorkerPool) { p.idleTimeout = d }
}

func WithQueueSize(size int) BoundedWorkerPoolOption {
	return func(p *BoundedWorkerPool) { p.queueSize = size }
}

func WithPanicHandler(handler PanicHandler) BoundedWorkerPoolOption {
	return func(p *BoundedWorkerPool) { p.panicHandler = handler }
}
