package ristretto

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type LocalCache[V any] struct {
	hub        *ristretto.Cache[string, any]
	defaultTTL time.Duration
}

func NewLocalCache[V any](
	cfg ristretto.Config[string, any],
	defaultTTL time.Duration,
) (*LocalCache[V], error) {
	cache, err := ristretto.NewCache(&cfg)
	if err != nil {
		return nil, err
	}
	return &LocalCache[V]{hub: cache, defaultTTL: defaultTTL}, nil
}

func (r *LocalCache[V]) Get(key string) (V, bool) {
	val, ok := r.hub.Get(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

func (r *LocalCache[V]) SetWithTTL(key string, value V, ttl time.Duration) {
	r.hub.SetWithTTL(key, value, 1, ttl)
}

func (r *LocalCache[V]) Set(key string, value V) {
	r.hub.SetWithTTL(key, value, 1, r.defaultTTL)
}

func (r *LocalCache[V]) Del(key string) {
	r.hub.Del(key)
}
