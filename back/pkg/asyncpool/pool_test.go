//go:build unit

package asyncpool_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/brice-74/sensorflow/pkg/asyncpool"
)

func waitUntil(t *testing.T, cond func() bool, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}

func TestSubmitAndExecution(t *testing.T) {
	var count atomic.Int64
	pool := asyncpool.NewWorkerPool(asyncpool.WithMinWorkers(1), asyncpool.WithMaxWorkers(4))
	defer pool.ForceShutdown()

	for range 10 {
		if err := pool.Submit(func() { count.Add(1) }); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	waitUntil(t, func() bool { return count.Load() == 10 }, 2*time.Second)
	if pool.QueueLength() != 0 {
		t.Errorf("expected queue empty, got %d", pool.QueueLength())
	}
}

func TestScalingBehavior(t *testing.T) {
	pool := asyncpool.NewWorkerPool(asyncpool.WithMinWorkers(1), asyncpool.WithMaxWorkers(5))
	defer pool.ForceShutdown()

	var wg sync.WaitGroup
	wg.Add(20)
	for range 20 {
		pool.Submit(func() {
			time.Sleep(100 * time.Millisecond)
			wg.Done()
		})
	}

	waitUntil(t, func() bool { return pool.ActiveWorkers() > 1 }, 1*time.Second)

	wg.Wait()
}

func TestPanicHandler(t *testing.T) {
	var recovered atomic.Bool
	pool := asyncpool.NewWorkerPool(asyncpool.WithPanicHandler(func(r any) {
		recovered.Store(true)
	}))
	defer pool.ForceShutdown()

	_ = pool.Submit(func() { panic("boom") })

	waitUntil(t, func() bool { return recovered.Load() }, 1*time.Second)
}

func TestShutdownGraceful(t *testing.T) {
	var count atomic.Int64
	pool := asyncpool.NewWorkerPool(asyncpool.WithMinWorkers(1), asyncpool.WithMaxWorkers(1))
	for range 5 {
		pool.Submit(func() {
			time.Sleep(50 * time.Millisecond)
			count.Add(1)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := pool.Shutdown(ctx)
	if err != nil {
		t.Fatalf("expected graceful shutdown, got %v", err)
	}

	if count.Load() != 5 {
		t.Errorf("expected 5 tasks executed, got %d", count.Load())
	}
}

func TestShutdownTimeout(t *testing.T) {
	pool := asyncpool.NewWorkerPool(asyncpool.WithMinWorkers(1), asyncpool.WithMaxWorkers(1))
	for range 2 {
		pool.Submit(func() { time.Sleep(300 * time.Millisecond) })
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := pool.Shutdown(ctx)
	if err != asyncpool.ErrShutdownTimeout {
		t.Fatalf("expected ErrShutdownTimeout, got %v", err)
	}
}

func TestForceShutdownDrainsQueue(t *testing.T) {
	var count atomic.Int64
	pool := asyncpool.NewWorkerPool(asyncpool.WithMinWorkers(1), asyncpool.WithMaxWorkers(1))

	var started sync.WaitGroup
	started.Add(1)

	err := pool.Submit(func() {
		started.Done()
		time.Sleep(200 * time.Millisecond)
		count.Add(1)
	})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	for range 5 {
		_ = pool.Submit(func() { count.Add(1) })
	}

	started.Wait()

	pool.ForceShutdown()

	final := count.Load()
	if final != 1 {
		t.Errorf("expected only the running task to complete (1), got %d", final)
	}
}

func TestSubmitAfterShutdown(t *testing.T) {
	pool := asyncpool.NewWorkerPool()
	_ = pool.Submit(func() {})
	_ = pool.Shutdown(context.Background())

	err := pool.Submit(func() {})
	if err != asyncpool.ErrQueueClosed {
		t.Errorf("expected ErrQueueClosed, got %v", err)
	}
}

func TestIdleWorkerExit(t *testing.T) {
	pool := asyncpool.NewWorkerPool(asyncpool.WithMinWorkers(0), asyncpool.WithMaxWorkers(2), asyncpool.WithIdleTimeout(100*time.Millisecond))
	defer pool.ForceShutdown()

	pool.Submit(func() {})
	waitUntil(t, func() bool { return pool.ActiveWorkers() > 0 }, 500*time.Millisecond)
	waitUntil(t, func() bool { return pool.ActiveWorkers() == 0 }, 2*time.Second)
}
