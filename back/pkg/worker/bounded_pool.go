package worker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type PoolCore[T Task] interface {
	Submit(task T) error
}

// PanicHandler is an optional callback for recovered panics in tasks.
type PanicHandler = func(any)

var (
	ErrQueueClosed     = errors.New("asyncpool queue closed")
	ErrQueueFull       = errors.New("asyncpool queue full")
	ErrShutdownTimeout = errors.New("asyncpool shutdown timeout")
)

// BoundedPool is a lightweight, adaptive worker pool for executing tasks asynchronously
// in order of submission (FIFO) and dynamically adjusts the number of workers.
type BoundedPool[T Task] struct {
	queue         chan T        // Channel holding pending tasks
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

var _ PoolCore[Task] = (*BoundedPool[Task])(nil)

func NewBoundedPool[T Task](opts ...BoundedPoolOption[T]) *BoundedPool[T] {
	p := &BoundedPool[T]{
		minWorkers:  1,
		maxWorkers:  10,
		idleTimeout: 5 * time.Second,
		queueSize:   1000,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.queue = make(chan T, p.queueSize)
	p.ctx, p.cancel = context.WithCancel(context.Background())

	for i := 0; i < p.minWorkers; i++ {
		p.spawnWorker()
	}

	return p
}

// Submit enqueues a task to the pool. Returns ErrQueueClosed or ErrQueueFull.
func (p *BoundedPool[T]) Submit(task T) error {
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
func (p *BoundedPool[T]) spawnWorker() {
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
func (p *BoundedPool[T]) tryScale() {
	qLen := len(p.queue)
	active := int(p.activeWorkers.Load())

	if qLen > 0 && (p.maxWorkers <= 0 || active < p.maxWorkers) {
		p.spawnWorker()
	}
}

// Shutdown gracefully stops the pool, waiting for all current and queued tasks to complete.
// Timeout can be provided via context.
func (p *BoundedPool[T]) Shutdown(ctx context.Context) error {
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
func (p *BoundedPool[T]) ForceShutdown() {
	if !p.isClosed.Swap(true) {
		close(p.queue)
		p.cancel()
	}

	for range p.queue {
	}
	p.wg.Wait()
}

// ActiveWorkers returns the current number of active workers.
func (p *BoundedPool[T]) ActiveWorkers() int64 {
	return p.activeWorkers.Load()
}

// QueueLength returns the current number of tasks waiting in the queue.
func (p *BoundedPool[T]) QueueLength() int {
	return len(p.queue)
}

type BoundedPoolOption[T Task] func(*BoundedPool[T])

func WithMinWorkers[T Task](min int) BoundedPoolOption[T] {
	return func(p *BoundedPool[T]) { p.minWorkers = min }
}

func WithMaxWorkers[T Task](max int) BoundedPoolOption[T] {
	return func(p *BoundedPool[T]) { p.maxWorkers = max }
}

func WithIdleTimeout[T Task](d time.Duration) BoundedPoolOption[T] {
	return func(p *BoundedPool[T]) { p.idleTimeout = d }
}

func WithQueueSize[T Task](size int) BoundedPoolOption[T] {
	return func(p *BoundedPool[T]) { p.queueSize = size }
}

func WithPanicHandler[T Task](handler PanicHandler) BoundedPoolOption[T] {
	return func(p *BoundedPool[T]) { p.panicHandler = handler }
}
