//go:build unit

package worker_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/brice-74/sensorflow/pkg/worker"
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
	p := worker.NewBoundedPool(worker.WithMinWorkers[worker.Task](1), worker.WithMaxWorkers[worker.Task](4))
	defer p.ForceShutdown()

	for i := 0; i < 10; i++ {
		if err := p.Submit(worker.VoidTask(func() { count.Add(1) })); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	waitUntil(t, func() bool { return count.Load() == 10 }, 2*time.Second)
	if p.QueueLength() != 0 {
		t.Errorf("expected queue empty, got %d", p.QueueLength())
	}
}

func TestScalingBehavior(t *testing.T) {
	p := worker.NewBoundedPool(worker.WithMinWorkers[worker.Task](1), worker.WithMaxWorkers[worker.Task](5))
	defer p.ForceShutdown()

	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		_ = p.Submit(worker.VoidTask(func() {
			time.Sleep(100 * time.Millisecond)
			wg.Done()
		}))
	}

	waitUntil(t, func() bool { return p.ActiveWorkers() > 1 }, 1*time.Second)

	wg.Wait()
}

func TestPanicHandler(t *testing.T) {
	var recovered atomic.Bool
	p := worker.NewBoundedPool(worker.WithPanicHandler[worker.Task](func(r any) {
		recovered.Store(true)
	}))
	defer p.ForceShutdown()

	_ = p.Submit(worker.VoidTask(func() { panic("boom") }))

	waitUntil(t, func() bool { return recovered.Load() }, 1*time.Second)
}

func TestShutdownGraceful(t *testing.T) {
	var count atomic.Int64
	p := worker.NewBoundedPool(worker.WithMinWorkers[worker.Task](1), worker.WithMaxWorkers[worker.Task](1))
	for i := 0; i < 5; i++ {
		_ = p.Submit(worker.VoidTask(func() {
			time.Sleep(50 * time.Millisecond)
			count.Add(1)
		}))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := p.Shutdown(ctx)
	if err != nil {
		t.Fatalf("expected graceful shutdown, got %v", err)
	}

	if count.Load() != 5 {
		t.Errorf("expected 5 tasks executed, got %d", count.Load())
	}
}

func TestShutdownTimeout(t *testing.T) {
	p := worker.NewBoundedPool(worker.WithMinWorkers[worker.Task](1), worker.WithMaxWorkers[worker.Task](1))
	for i := 0; i < 2; i++ {
		_ = p.Submit(worker.VoidTask(func() { time.Sleep(300 * time.Millisecond) }))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := p.Shutdown(ctx)
	if err != worker.ErrShutdownTimeout {
		t.Fatalf("expected ErrShutdownTimeout, got %v", err)
	}
}

func TestForceShutdownDrainsQueue(t *testing.T) {
	var count atomic.Int64
	p := worker.NewBoundedPool(worker.WithMinWorkers[worker.Task](1), worker.WithMaxWorkers[worker.Task](1))

	var started sync.WaitGroup
	started.Add(1)

	err := p.Submit(worker.VoidTask(func() {
		started.Done()
		time.Sleep(200 * time.Millisecond)
		count.Add(1)
	}))
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	for i := 0; i < 5; i++ {
		_ = p.Submit(worker.VoidTask(func() { count.Add(1) }))
	}

	started.Wait()

	p.ForceShutdown()

	final := count.Load()
	if final != 1 {
		t.Errorf("expected only the running task to complete (1), got %d", final)
	}
}

func TestSubmitAfterShutdown(t *testing.T) {
	p := worker.NewBoundedPool[worker.Task]()
	_ = p.Submit(worker.VoidTask(func() {}))
	_ = p.Shutdown(context.Background())

	err := p.Submit(worker.VoidTask(func() {}))
	if err != worker.ErrQueueClosed {
		t.Errorf("expected ErrQueueClosed, got %v", err)
	}
}

func TestIdleWorkerExit(t *testing.T) {
	p := worker.NewBoundedPool(worker.WithMinWorkers[worker.Task](0), worker.WithMaxWorkers[worker.Task](2), worker.WithIdleTimeout[worker.Task](100*time.Millisecond))
	defer p.ForceShutdown()
	_ = p.Submit(worker.VoidTask(func() {}))
	waitUntil(t, func() bool { return p.ActiveWorkers() > 0 }, 500*time.Millisecond)
	waitUntil(t, func() bool { return p.ActiveWorkers() == 0 }, 2*time.Second)
}
