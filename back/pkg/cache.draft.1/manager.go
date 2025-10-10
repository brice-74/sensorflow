package cache

import (
	"context"
	"errors"
	"time"
)

// CacheManager coordonne Local, Distributed et Invalidator.
type CacheManager[K comparable, V any] struct {
	Local       Local[K, V]
	Distributed Distributed[K, V]
	Invalidator Invalidator[K]
	TTL         time.Duration
}

// NewCacheManager enregistre le handler d’invalidation local si besoin.
func NewCacheManager[K comparable, V any](
	ctx context.Context,
	local Local[K, V],
	distributed Distributed[K, V],
	invalidator Invalidator[K],
	ttl time.Duration,
) (*CacheManager[K, V], error) {

	m := &CacheManager[K, V]{
		Local:       local,
		Distributed: distributed,
		Invalidator: invalidator,
		TTL:         ttl,
	}

	if err := invalidator.SubscribeInvalidations(func(ctx context.Context, key K) error {
		m.Local.Del(key)
		return nil
	}); err != nil {
		return nil, err
	}

	// On s’assure que le listener Redis tourne (Start est idempotent)
	if err := invalidator.Start(ctx); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *CacheManager[K, V]) Get(ctx context.Context, key K) (V, bool, error) {
	if val, ok := m.Local.Get(key); ok {
		return val, true, nil
	}

	val, ok, err := m.Distributed.Get(ctx, key)
	if err != nil || !ok {
		var zero V
		return zero, false, err
	}

	m.Local.Set(key, val, m.TTL)
	return val, true, nil
}

func (m *CacheManager[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	m.Local.Set(key, value, ttl)
	if err := m.Distributed.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	return m.Invalidator.PublishInvalidate(ctx, key)
}

func (m *CacheManager[K, V]) Invalidate(ctx context.Context, key K) error {
	m.Local.Del(key)

	return m.Invalidator.PublishInvalidate(ctx, key)

}

func (m *CacheManager[K, V]) Close(ctx context.Context) error {
	var errs []error
	errs = append(errs, m.Invalidator.Close(ctx))
	errs = append(errs, m.Distributed.Close(ctx))
	m.Local.Close()
	return errors.Join(errs...)
}
