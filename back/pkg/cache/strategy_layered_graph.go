package cache

import (
	"context"
	"time"
)

type LayeredGraphStrategy[K comparable, V any] struct {
	local       LocalCache[K, V]
	distributed DistributedCache[K, V]
	invalidator Invalidator[K]
	ttl         time.Duration
}

func NewLayeredGraphStrategy[K comparable, V any](
	local LocalCache[K, V],
	distributed DistributedCache[K, V],
	invalidator Invalidator[K],
	ttl time.Duration,
) *LayeredGraphStrategy[K, V] {
	return &LayeredGraphStrategy[K, V]{
		local:       local,
		distributed: distributed,
		invalidator: invalidator,
		ttl:         ttl,
	}
}

func (s *LayeredGraphStrategy[K, V]) Get(ctx context.Context, key K, fallback func() (V, error)) (V, error) {
	if val, ok := s.local.Get(key); ok {
		return val, nil
	}

	val, err := s.distributed.Get(ctx, key)
	if err == nil && any(val) != nil {
		s.local.Set(key, val)
		return val, nil
	}

	val, err = fallback()
	if err != nil {
		return val, err
	}

	s.local.Set(key, val)
	return val, nil
}

func (s *LayeredGraphStrategy[K, V]) MGet(ctx context.Context, keys []K, fallback func() (map[K]V, error)) (map[K]V, error) {
	// TODO: Implement SMEMBERS + MGET pipeline version
	return nil, nil
}

func (s *LayeredGraphStrategy[K, V]) Set(ctx context.Context, key K, value V) error {
	s.local.Set(key, value)
	return s.distributed.Set(ctx, key, value)
}

func (s *LayeredGraphStrategy[K, V]) SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error {
	s.local.SetWithTTL(key, value, ttl)
	return s.distributed.SetWithTTL(ctx, key, value, ttl)
}

func (s *LayeredGraphStrategy[K, V]) Delete(ctx context.Context, key K) error {
	s.local.Delete(key)
	return s.distributed.Delete(ctx, key)
}

func (s *LayeredGraphStrategy[K, V]) Invalidate(ctx context.Context, keys []K) error {
	for _, key := range keys {
		s.local.Delete(key)
	}
	return s.invalidator.PublishInvalidate(ctx, keys)
}
