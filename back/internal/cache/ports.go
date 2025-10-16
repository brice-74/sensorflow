package cache

import (
	"context"
	"time"
)

// Local represents a local memory cache (Ristretto, Freecache, etc.)
type Local[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	SetWithTTL(key K, value V, ttl time.Duration)
	Delete(key K)
}

// Invalidator allows invalidations to be propagated across the network.
type Invalidator[K comparable] interface {
	PublishEvict(ctx context.Context, keys []K) error
	SubscribeEvict(ctx context.Context, handler func(keys []K))
}
