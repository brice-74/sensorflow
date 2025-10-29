package redis

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// HealthyClient wraps a redis.Client and tracks its health.
// It marks itself down on request errors and tries to recover using configurable exponential backoff with optional jitter.
type HealthyClient struct {
	Client         *redis.Client
	isDown         atomic.Bool
	pingTimeout    time.Duration
	initialBackoff time.Duration
	maxBackoff     time.Duration
	multiplier     float64
	jitterPct      float64
}

type Option func(*HealthyClient)

func WithPingTimeout(timeout time.Duration) Option {
	return func(h *HealthyClient) {
		h.pingTimeout = timeout
	}
}

func WithBackoff(initial, max time.Duration, multiplier, jitterPct float64) Option {
	return func(h *HealthyClient) {
		h.initialBackoff = initial
		h.maxBackoff = max
		h.multiplier = multiplier
		h.jitterPct = jitterPct
	}
}

func NewHealthyClient(client *redis.Client, opts ...Option) *HealthyClient {
	h := &HealthyClient{
		Client:         client,
		pingTimeout:    500 * time.Millisecond,
		initialBackoff: 100 * time.Millisecond,
		maxBackoff:     30 * time.Second,
		multiplier:     2,
		jitterPct:      0.0, // default: no jitter
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *HealthyClient) HandleError(err error) error {
	return HandleError(err, h.markDown)
}

// markDown sets the client as down and triggers recovery if previously healthy.
func (h *HealthyClient) markDown() {
	if !h.isDown.Swap(true) {
		go h.tryRecover()
	}
}

// tryRecover pings Redis periodically with exponential backoff and optional jitter until recovery.
// How jitter works: if the calculated backoff is 1 s, then with 0.1 the actual delay will be randomly selected between 0.9 s and 1.1 s.
func (h *HealthyClient) tryRecover() {
	backoff := h.initialBackoff

	for h.isDown.Load() {
		ctx, cancel := context.WithTimeout(context.Background(), h.pingTimeout)
		err := h.Client.Ping(ctx).Err()
		cancel()

		if err == nil {
			h.isDown.Store(false)
			return
		}

		sleep := backoff
		if h.jitterPct > 0 {
			factor := 1 + (rand.Float64()*2-1)*h.jitterPct
			sleep = time.Duration(float64(sleep) * factor)
		}

		time.Sleep(sleep)

		next := min(time.Duration(float64(backoff)*h.multiplier), h.maxBackoff)
		backoff = next
	}
}

// IsHealthy returns true if the client is currently healthy.
func (h *HealthyClient) IsHealthy() bool {
	return !h.isDown.Load()
}
