package cache

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type Ristretto[K comparable, V any] struct {
	hub     *ristretto.Cache[string, any]
	prefix  string
	keyFunc func(K) string
}

func NewRistretto[K comparable, V any](
	cfg ristretto.Config[string, any],
	prefix string,
	keyFunc func(K) string,
) (*Ristretto[K, V], error) {
	cache, err := ristretto.NewCache(&cfg)
	if err != nil {
		return nil, err
	}
	return &Ristretto[K, V]{hub: cache, prefix: prefix, keyFunc: keyFunc}, nil
}

func (r *Ristretto[K, V]) makeKey(k K) string {
	return fmt.Sprintf("%s:%v", r.prefix, r.keyFunc(k))
}

func (r *Ristretto[K, V]) Get(key K) (V, bool) {
	val, ok := r.hub.Get(r.makeKey(key))
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

func (r *Ristretto[K, V]) Set(key K, value V, ttl time.Duration) {
	r.hub.SetWithTTL(r.makeKey(key), value, 1, ttl)
}

func (r *Ristretto[K, V]) Del(key K) {
	r.hub.Del(r.makeKey(key))
}
