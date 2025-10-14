package ristretto

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type LocalCache[K comparable, V any] struct {
	hub       *ristretto.Cache[string, any]
	formatKey func(K) string
	ttl       time.Duration
}

func NewLocalCache[K comparable, V any](
	cfg ristretto.Config[string, any],
	prefix string,
	formatKey func(K) string,
	defaultTTL time.Duration,
) (*LocalCache[K, V], error) {
	cache, err := ristretto.NewCache(&cfg)
	if err != nil {
		return nil, err
	}
	return &LocalCache[K, V]{hub: cache, formatKey: formatKey, ttl: defaultTTL}, nil
}

func (r *LocalCache[K, V]) Get(key K) (V, bool) {
	val, ok := r.hub.Get(r.formatKey(key))
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

func (r *LocalCache[K, V]) Set(key K, value V) {
	r.hub.SetWithTTL(r.formatKey(key), value, 1, r.ttl)
}

func (r *LocalCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	r.hub.SetWithTTL(r.formatKey(key), value, 1, ttl)
}

func (r *LocalCache[K, V]) Del(key K) {
	r.hub.Del(r.formatKey(key))
}
