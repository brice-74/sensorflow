package factory

import (
	"time"
)

type Factory struct {
	now func() time.Time
}

type Option func(*Factory)

func New(opts ...Option) *Factory {
	f := &Factory{now: time.Now().UTC}
	for _, opt := range opts {
		opt(f)
	}
	if f.now == nil {
		f.now = time.Now().UTC
	}
	return f
}

func WithNow(now func() time.Time) Option {
	return func(f *Factory) {
		f.now = now
	}
}
