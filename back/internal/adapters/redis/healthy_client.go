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
	*redis.Client
	isDown atomic.Bool
	opt    *Options
}

var _ Client = (*HealthyClient)(nil)

type Options struct {
	PingTimeout    time.Duration
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	JitterPct      float64
}

func (opts *Options) init() {
	if opts.PingTimeout == 0 {
		opts.PingTimeout = 500 * time.Millisecond
	}
	if opts.InitialBackoff == 0 {
		opts.InitialBackoff = 100 * time.Millisecond
	}
	if opts.MaxBackoff == 0 {
		opts.MaxBackoff = 30 * time.Second
	}
	if opts.Multiplier == 0 {
		opts.Multiplier = 2
	}
}

func NewHealthyClient(client *redis.Client, opts *Options) *HealthyClient {
	if opts == nil {
		panic("redis: NewHealthyClient nil options")
	}
	opts.init()

	return &HealthyClient{
		Client: client,
		opt:    opts,
	}
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
	backoff := h.opt.InitialBackoff

	for h.isDown.Load() {
		ctx, cancel := context.WithTimeout(context.Background(), h.opt.PingTimeout)
		err := h.Client.Ping(ctx).Err()
		cancel()

		if err == nil {
			h.isDown.Store(false)
			return
		}

		sleep := backoff
		if h.opt.JitterPct > 0 {
			factor := 1 + (rand.Float64()*2-1)*h.opt.JitterPct
			sleep = time.Duration(float64(sleep) * factor)
		}

		time.Sleep(sleep)

		next := min(time.Duration(float64(backoff)*h.opt.Multiplier), h.opt.MaxBackoff)
		backoff = next
	}
}

// IsHealthy returns true if the client is currently healthy.
func (h *HealthyClient) IsHealthy() bool {
	return !h.isDown.Load()
}

func (h *HealthyClient) Cmdable(ctx context.Context) redis.Cmdable {
	return h.Client
}

func (h *HealthyClient) NewPipeline(ctx context.Context) (redis.Pipeliner, context.Context) {
	pipe := h.Client.Pipeline()
	return pipe, WithPipeline(ctx, pipe)
}

func (h *HealthyClient) NewConn(ctx context.Context) (redis.Cmdable, context.Context) {
	conn := h.Client.Conn()
	return conn, WithConn(ctx, conn)
}
