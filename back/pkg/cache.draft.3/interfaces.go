package cache

import (
	"context"
	"time"
)

// LocalCache represents a local memory cache (Ristretto, Freecache, etc.)
type LocalCache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	SetWithTTL(key K, value V, ttl time.Duration)
	Delete(key K)
}

// DistributedCache represents a distributed cache (Redis, Memcached, etc.)
type DistributedCache[K comparable, V any] interface {
	Get(ctx context.Context, key K) (V, error)
	MGet(ctx context.Context, keys ...K) (map[K]V, error)
	Set(ctx context.Context, key K, value V) error
	SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error
	Delete(ctx context.Context, key K) error
}

// Invalidator allows invalidations to be propagated across the network.
type Invalidator[K comparable] interface {
	PublishEvict(ctx context.Context, keys []K) error
	SubscribeEvict(ctx context.Context, handler func(keys []K))
}

// Manager defines how the cache is orchestrated (local + distributed).
type Manager[K comparable, V any] interface {
	Get(ctx context.Context, key K, fallback func() (V, error)) (V, error)
	MGet(ctx context.Context, keys []K, fallback func() (map[K]V, error)) (map[K]V, error)
	MGetGraph(ctx context.Context, plan FetchPlan[K], fallback func([]K) (map[K]V, error)) (map[K]V, error)
	Set(ctx context.Context, key K, value V, ttl time.Duration) error
	SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error
	Delete(ctx context.Context, key K) error
	Evict(ctx context.Context, keys []K) error
}

type FetchPlan[K comparable] struct {
}
