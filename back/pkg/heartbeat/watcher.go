package heartbeat

import (
	"context"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

type Options struct {
	RecoverTimeout time.Duration
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
	isAliveCh   chan bool
	cancel      context.CancelFunc
}

func NewWatcher[T AliveCapable](client T, handleError func(error, func()) error, opts *Options, useIsAliveCh bool) *Watcher[T] {
	if opts == nil {
		panic("heartbeat: Watcher nil options")
	}
	opts.init()

	var ch chan bool
	if useIsAliveCh {
		ch = make(chan bool, 1)
	}

	return &Watcher[T]{
		Client:      client,
		opt:         opts,
		handleError: handleError,
		isAliveCh:   ch,
	}
}

func (l *Watcher[T]) HandleError(err error) error {
	return l.handleError(err, l.MarkDown)
}

func (l *Watcher[T]) CancelRecover() {
	if l.cancel != nil {
		l.cancel()
	}
}

func (l *Watcher[T]) MarkDown() {
	if !l.isDown.Swap(true) {
		l.emit(false)
		ctx, cancel := context.WithCancel(context.Background())
		l.cancel = cancel
		go l.tryRecover(ctx)
	}
}

func (l *Watcher[T]) tryRecover(ctx context.Context) {
	backoff := l.opt.InitialBackoff

	var recoverCtx context.Context
	var cancel context.CancelFunc

	if l.opt.RecoverTimeout > 0 {
		recoverCtx, cancel = context.WithTimeout(ctx, l.opt.RecoverTimeout)
		defer cancel()
	} else {
		recoverCtx = ctx
	}

	for l.isDown.Load() {
		select {
		case <-recoverCtx.Done():
			return
		default:
		}

		cctx, cancelPing := context.WithTimeout(recoverCtx, l.opt.PingTimeout)
		alive := l.Client.Alive(cctx)
		cancelPing()

		if alive {
			l.isDown.Store(false)
			l.emit(true)
			return
		}

		sleep := backoff
		if l.opt.JitterPct > 0 {
			factor := 1 + (rand.Float64()*2-1)*l.opt.JitterPct
			sleep = time.Duration(float64(sleep) * factor)
		}

		select {
		case <-recoverCtx.Done():
			return
		case <-time.After(sleep):
		}

		backoff = min(time.Duration(float64(backoff)*l.opt.Multiplier), l.opt.MaxBackoff)
	}
}

func (l *Watcher[T]) IsAlive() bool {
	return !l.isDown.Load()
}

func (l *Watcher[T]) IsAliveCh() <-chan bool {
	return l.isAliveCh
}

func (l *Watcher[T]) emit(alive bool) {
	if l.isAliveCh == nil {
		return
	}
	select {
	case <-l.isAliveCh:
	default:
	}
	select {
	case l.isAliveCh <- alive:
	default:
	}
}
