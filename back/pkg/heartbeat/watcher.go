package heartbeat

import (
	"context"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

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

type AliveCapable interface {
	Alive(ctx context.Context) bool
}

type Watcher[T AliveCapable] struct {
	Client      T
	isDown      atomic.Bool
	opt         *Options
	handleError func(error, func()) error
}

func NewWatcher[T AliveCapable](client T, handleError func(error, func()) error, opts *Options) *Watcher[T] {
	if opts == nil {
		panic("heartbeat: Watcher nil options")
	}
	opts.init()

	return &Watcher[T]{
		Client:      client,
		opt:         opts,
		handleError: handleError,
	}
}

func (l *Watcher[T]) HandleError(err error) error {
	return l.handleError(err, l.MarkDown)
}

func (l *Watcher[T]) MarkDown() {
	if !l.isDown.Swap(true) {
		go l.tryRecover()
	}
}

func (l *Watcher[T]) tryRecover() {
	backoff := l.opt.InitialBackoff

	for l.isDown.Load() {
		ctx, cancel := context.WithTimeout(context.Background(), l.opt.PingTimeout)
		alive := l.Client.Alive(ctx)
		cancel()

		if alive {
			l.isDown.Store(false)
			return
		}

		sleep := backoff
		if l.opt.JitterPct > 0 {
			factor := 1 + (rand.Float64()*2-1)*l.opt.JitterPct
			sleep = time.Duration(float64(sleep) * factor)
		}

		time.Sleep(sleep)
		backoff = min(time.Duration(float64(backoff)*l.opt.Multiplier), l.opt.MaxBackoff)
	}
}

func (l *Watcher[T]) IsAlive() bool {
	return !l.isDown.Load()
}
