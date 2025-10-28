package asyncpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type (
	// Task represents a unit of work to be executed by the pool.
	Task func()
	// PanicHandler is an optional callback for recovered panics in tasks.
	PanicHandler func(any)
)

var (
	ErrQueueClosed     = errors.New("asyncpool queue closed")
	ErrQueueFull       = errors.New("asyncpool queue full")
	ErrShutdownTimeout = errors.New("asyncpool shutdown timeout")
)

// WorkerPool is a lightweight, adaptive worker pool for executing tasks asynchronously.
type WorkerPool struct {
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
	mu       sync.Mutex
	isClosed atomic.Bool
}

func NewWorkerPool(opts ...Option) *WorkerPool {
	p := &WorkerPool{
		minWorkers:  0,
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
func (p *WorkerPool) Submit(task Task) error {
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
func (p *WorkerPool) spawnWorker() {
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
					task()
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
func (p *WorkerPool) tryScale() {
	qLen := len(p.queue)
	active := int(p.activeWorkers.Load())

	if qLen > 0 && (p.maxWorkers <= 0 || active < p.maxWorkers) {
		p.spawnWorker()
	}
}

// Shutdown gracefully stops the pool, waiting for all current and queued tasks to complete.
// Timeout can be provided via context.
func (p *WorkerPool) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	if !p.isClosed.Swap(true) {
		close(p.queue)
	}
	p.mu.Unlock()

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
func (p *WorkerPool) ForceShutdown() {
	p.mu.Lock()
	if !p.isClosed.Swap(true) {
		close(p.queue)
		p.cancel()
	}
	p.mu.Unlock()

	for range p.queue {
	}
	p.wg.Wait()
}

// ActiveWorkers returns the current number of active workers.
func (p *WorkerPool) ActiveWorkers() int64 {
	return p.activeWorkers.Load()
}

// QueueLength returns the current number of tasks waiting in the queue.
func (p *WorkerPool) QueueLength() int {
	return len(p.queue)
}
